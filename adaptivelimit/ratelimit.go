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

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var ErrLimit = "Hertz Adaptlive limiting"

/*
	CPU sampling algorithm using BBR
*/
func Adaptlivelimit(opts ...Option) app.HandlerFunc {
	limiter := NewLimiter(opts...)
	return func(c context.Context, ctx *app.RequestContext) {
		done, err := limiter.Allow()
		if err != nil {
			ctx.String(consts.StatusTooManyRequests, ErrLimit)
			ctx.Abort()
		} else {
			ctx.Next(c)
			done()
		}
	}
}
