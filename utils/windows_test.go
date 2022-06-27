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

package utils

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRollingWindow(t *testing.T) {
	assert.NotNil(t, NewRollingWindow(10, time.Second))
	assert.Panics(t, func() {
		NewRollingWindow(0, time.Second)
	})
}

func TestRollingWindowAdd(t *testing.T) {
	r := NewRollingWindow(3, time.Millisecond*5)
	list := func() []float64 {
		var buckets []float64
		r.Reduce(func(b *Bucket) {
			buckets = append(buckets, b.Sum)
		})
		return buckets
	}
	assert.Equal(t, []float64{0, 0, 0}, list())
	r.Add(1)
	assert.Equal(t, []float64{0, 0, 1}, list())
	time.Sleep(5 * time.Millisecond)
	// next cycle
	r.Add(2)
	r.Add(3)
	//  0 0 1  -> 0 1 0 -> 0 1 2 -> 0 1 5
	assert.Equal(t, []float64{0, 1, 5}, list())
}

func TestRollingWindowSum(t *testing.T) {
	r := NewRollingWindow(3, time.Millisecond*5)
	var cnt float64
	list := func() float64 {
		r.Reduce(func(b *Bucket) {
			cnt = math.Max(cnt, b.Sum)
		})
		return cnt
	}
	assert.Equal(t, float64(0), list())
	r.Add(1)
	assert.Equal(t, float64(1), list())
	time.Sleep(5 * time.Millisecond)
	// next cycle
	r.Add(2)
	r.Add(3)
	//  0 0 1  -> 0 1 0 -> 0 1 2 -> 0 1 5
	assert.Equal(t, float64(5), list())
}

func TestRollingWindowsAvg(t *testing.T) {
	r := NewRollingWindow(3, time.Second*5)
	var cnt float64 = 1 << 31
	list := func() float64 {
		r.Reduce(func(b *Bucket) {
			if b.Count <= 0 {
				return
			}
			if cnt > math.Ceil(b.Sum/float64(b.Count)) {
				cnt = math.Ceil(b.Sum / float64(b.Count))
			}
		})
		if cnt == 1<<31 {
			return 1
		}
		return cnt
	}
	assert.Equal(t, float64(1), list())
	r.Add(1)
	assert.Equal(t, float64(1), list())
	time.Sleep(5 * time.Second)
	// next cycle
	r.Add(2)
	r.Add(3)
	time.Sleep(5 * time.Second)
	r.Add(4)
	r.Add(5)
	assert.Equal(t, float64(1), list())
}
