/*
 * @Author: lvhaidong
 * @Date: 2022-06-19 23:26:28
 * @LastEditors: lvhaidong
 * @LastEditTime: 2022-06-20 08:39:49
 * @Description:
 */
package adaptivelimit

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

var (
	ErrLimit = "Hertz Rate limiting"
)

/*
	CPU sampling algorithm using BBR
*/
func Ratelimit(opts ...options) app.HandlerFunc {
	limiter := NewLimiter()
	return func(c context.Context, ctx *app.RequestContext) {
		done, err := limiter.Allow()
		if err != nil {
			ctx.String(consts.StatusTooManyRequests, ErrLimit)
			return
		}
		ctx.Next(c)
		done()
	}
}
