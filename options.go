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

import "time"

type Option func(o *options)

var opt = options{
	Window:       time.Second * 10,
	Bucket:       100,                    // 100ms
	CPUThreshold: 800,                    // CPU load  80%
	SamplingTime: 500 * time.Millisecond, //
	Decay:        0.95,                   //
}

type options struct {
	Window       time.Duration
	Bucket       int
	CPUThreshold int64
	SamplingTime time.Duration
	Decay        float64
}

/*
	WithWindow: defines time duration per window
*/
func WithWindow(window time.Duration) Option {
	return func(o *options) {
		o.Window = window
	}
}

/*
	WithBucket: defines bucket number for each window
*/
func WithBucket(bucket int) Option {
	return func(o *options) {
		o.Bucket = bucket
	}
}

/*
	WithCPUThreshold: defines cpu threshold load cputhreshold / 1000
*/
func WithCPUThreshold(threshold int64) Option {
	return func(o *options) {
		o.CPUThreshold = threshold
	}
}

/*
	WithSamplingTime: defines cpu sampling time interval
*/
func WithSamplingTime(samplingTime time.Duration) Option {
	return func(o *options) {
		o.SamplingTime = samplingTime
	}
}

/*
	WithDecay: defines cpu attenuation factor
*/
func WithDecay(decay float64) Option {
	return func(o *options) {
		o.Decay = decay
	}
}

func NewOption(opts ...Option) options {
	for _, apply := range opts {
		apply(&opt)
	}
	return opt
}
