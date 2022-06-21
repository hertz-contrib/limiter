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

package adaptivelimit

import "time"

type Option func(o *options)

// default option
var opt = options{
	Window:       time.Second * 10,
	Bucket:       100,                    // 100ms
	CPUThreshold: 800,                    // CPU load  80%
	SamplingTime: 500 * time.Millisecond, //
	Decay:        0.95,                   //
}

// options of bbr limiter.
type options struct {
	// WindowSize defines time duration per window
	Window time.Duration
	// BucketNum defines bucket number for each window
	Bucket int
	// CPUThreshold
	CPUThreshold int64
	SamplingTime time.Duration
	Decay        float64
}

//  window size.
func WithWindow(window time.Duration) Option {
	return func(o *options) {
		o.Window = window
	}
}

// bucket ize.
func WithBucket(bucket int) Option {
	return func(o *options) {
		o.Bucket = bucket
	}
}

// cpu threshold
func WithCPUThreshold(threshold int64) Option {
	return func(o *options) {
		o.CPUThreshold = threshold
	}
}

// sapmleing time
func WithSamplingTime(samplingTime time.Duration) Option {
	return func(o *options) {
		o.SamplingTime = samplingTime
	}
}

// decay time
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
