/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package limiter

import (
	"errors"
	"math"
	"sync/atomic"
	"time"

	"github.com/c9s/goprocinfo/linux"
	"github.com/hertz-contrib/limiter/utils"
)

var (
	gCPU     int64
	gStat    linux.CPUStat
	ErrLimit = "Hertz Adaptlive limit"
)

type (
	cpuGetter func() int64
)

/*
	getCpuLoad : Get CPU State by reading /proc/stat
*/
func getCpuLoad() linux.CPUStat {
	stat, err := linux.ReadStat("/proc/stat")
	if err != nil {
		panic("stat read fail")
	}
	return stat.CPUStatAll
}

/*
	calcCoreUsage ：Calculate the overall utilization by reading the previous CPU state and the current CPU state
*/
func calcCoreUsage(curr, prev linux.CPUStat) float64 {
	PrevIdle := prev.Idle + prev.IOWait
	Idle := curr.Idle + curr.IOWait

	PrevNonIdle := prev.User + prev.Nice + prev.System + prev.IRQ + prev.SoftIRQ + prev.Steal
	NonIdle := curr.User + curr.Nice + curr.System + curr.IRQ + curr.SoftIRQ + curr.Steal

	PrevTotal := PrevIdle + PrevNonIdle
	Total := Idle + NonIdle
	totald := Total - PrevTotal
	idled := Idle - PrevIdle

	CPU_Percentage := (float64(totald) - float64(idled)) / float64(totald)

	return CPU_Percentage
}

func init() {
	go cpuProc()
}

/*
	cpuProc : CPU load correction by EMA algorithm
*/
func cpuProc() {
	ticker := time.NewTicker(opt.SamplingTime) // same to cpu sample rate
	defer func() {
		ticker.Stop()
		if err := recover(); err != nil {
			go cpuProc()
		}
	}()

	// EMA algorithm: https://blog.csdn.net/m0_38106113/article/details/81542863
	for range ticker.C {
		preState := gStat
		curState := getCpuLoad()
		usage := calcCoreUsage(preState, curState)
		prevCPU := atomic.LoadInt64(&gCPU)
		curCPU := int64(float64(prevCPU)*opt.Decay + float64(usage*10)*(1.0-opt.Decay))
		atomic.StoreInt64(&gCPU, curCPU)
	}
}

/*
	counterCache is used to cache maxPASS and minRt result.
*/
type counterCache struct {
	val  int64
	time time.Time
}

/*
	BBR implements bbr-like limiter.
	It is inspired by sentinel.
	https://github.com/alibaba/Sentinel/wiki/%E7%B3%BB%E7%BB%9F%E8%87%AA%E9%80%82%E5%BA%94%E9%99%90%E6%B5%81
*/
type BBR struct {
	cpu             cpuGetter
	passStat        *utils.RollingWindow // request succeeded
	rtStat          *utils.RollingWindow // time consume
	inFlight        int64                // Number of requests being processed
	bucketPerSecond int64
	bucketDuration  time.Duration

	// prevDropTime defines previous start drop since initTime
	prevDropTime atomic.Value
	maxPASSCache atomic.Value
	minRtCache   atomic.Value

	opts options
}

func NewLimiter(opts ...Option) *BBR {
	opt := NewOption(opts...)
	bucketDuration := opt.Window / time.Duration(opt.Bucket)
	// 10s / 100  = 100ms
	passStat := utils.NewRollingWindow(opt.Bucket, bucketDuration, utils.IgnoreCurrentBucket())
	rtStat := utils.NewRollingWindow(opt.Bucket, bucketDuration, utils.IgnoreCurrentBucket())

	limiter := &BBR{
		opts:            opt,
		passStat:        passStat,
		rtStat:          rtStat,
		bucketDuration:  bucketDuration,
		bucketPerSecond: int64(time.Second / bucketDuration),
		cpu:             func() int64 { return atomic.LoadInt64(&gCPU) },
	}

	return limiter
}

/*
	maxPass: Maximum number of requests in a single sampling window
*/
func (l *BBR) maxPass() int64 {
	passCache := l.maxPASSCache.Load()
	if passCache != nil {
		ps := passCache.(*counterCache)
		if l.timespan(ps.time) < 1 {
			return ps.val
		}
		// Avoid glitches caused by fluctuations
	}
	var rawMaxPass float64
	l.passStat.Reduce(func(b *utils.Bucket) {
		rawMaxPass = math.Max(float64(b.Sum), rawMaxPass)
	})
	if rawMaxPass <= 0 {
		rawMaxPass = 1
	}
	l.maxPASSCache.Store(&counterCache{
		val:  int64(rawMaxPass),
		time: time.Now(),
	})
	return int64(rawMaxPass)
}

/*
	timespan: returns the passed bucket count
*/
func (l *BBR) timespan(lastTime time.Time) int {
	v := int(time.Since(lastTime) / l.bucketDuration)
	if v > -1 {
		return v
	}
	return l.opts.Bucket
}

/*
	minRT: Minimum response time
*/
func (l *BBR) minRT() int64 {
	rtCache := l.minRtCache.Load()
	if rtCache != nil {
		rc := rtCache.(*counterCache)
		if l.timespan(rc.time) < 1 {
			return rc.val
		}
	}
	// Go to the nearest response time within 1s
	var rawMinRT float64 = 1 << 31
	l.rtStat.Reduce(func(b *utils.Bucket) {
		if b.Count <= 0 {
			return
		}
		if rawMinRT > math.Ceil(b.Sum/float64(b.Count)) {
			rawMinRT = math.Ceil(b.Sum / float64(b.Count))
		}
	})
	if rawMinRT == 1<<31 {
		rawMinRT = 1
	}
	l.minRtCache.Store(&counterCache{
		val:  int64(rawMinRT),
		time: time.Now(),
	})
	return int64(rawMinRT)
}

/*
	maxInFlight: Calculating the load
*/
func (l *BBR) maxInFlight() int64 {
	return int64(math.Ceil(float64(l.maxPass()*l.minRT()*l.bucketPerSecond) / 1000.0))
}

/*
	shouldDrop：(CPU load > 80% || (now - prevDrop) < 1s) and (MaxPass * MinRT * windows) / 1000 < InFlight
*/
func (l *BBR) shouldDrop() bool {
	now := time.Duration(time.Now().UnixNano())
	if l.cpu() < l.opts.CPUThreshold {
		// current cpu payload below the threshold
		prevDropTime, _ := l.prevDropTime.Load().(time.Duration)
		if prevDropTime == 0 {
			// haven't start drop,
			// accept current request
			return false
		}
		if time.Duration(now-prevDropTime) <= time.Second {
			// just start drop one second ago,
			// check current inflight count
			inFlight := atomic.LoadInt64(&l.inFlight)
			return inFlight > 1 && inFlight > l.maxInFlight()
		}
		l.prevDropTime.Store(time.Duration(0))
		return false
	}
	// current cpu payload exceeds the threshold
	inFlight := atomic.LoadInt64(&l.inFlight)
	drop := inFlight > 1 && inFlight > l.maxInFlight()
	if drop {
		prevDrop, _ := l.prevDropTime.Load().(time.Duration)
		if prevDrop != 0 {
			// already started drop, return directly
			return drop
		}
		// store start drop time
		l.prevDropTime.Store(now)
	}
	return drop
}

/*
	Allow：Determine the alarm triggering conditions, record the interface time consumption and QPS
*/
func (l *BBR) Allow() (func(), error) {
	if l.shouldDrop() {
		return nil, errors.New(ErrLimit)
	}
	atomic.AddInt64(&l.inFlight, 1)
	start := time.Now().UnixNano()
	// DoneFunc record time-consuming
	return func() {
		rt := (time.Now().UnixNano() - start) / int64(time.Millisecond)
		l.rtStat.Add(float64(rt))
		atomic.AddInt64(&l.inFlight, -1)
		l.passStat.Add(1)
	}, nil
}
