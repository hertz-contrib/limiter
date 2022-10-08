# limiter (This is a community driven project)

###  Adaptive Algorithm for [Hertz](https://github.com/cloudwego/hertz)

This project is inspired by kratos [limiter](https://github.com/go-kratos/aegis/blob/main/ratelimit/README.md). Thanks to their project.

#### Algorithm core

&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;The CPU load is collected by reading /proc/stat, and the CPU load is judged to trigger the flow restriction condition.
-  When the CPU load is less than the expected value: the current time is less than 1s from the last trigger to limit the flow, then determine whether the current maximum number of requests is greater than the past maximum load situation, if it is greater than the load situation, then limit the flow.
-  When the CPU load is greater than the expected value: determine whether the current number of requests is greater than the maximum load in the past, if it is greater than the maximum load in the past, then flow restriction will be performed.

#### How to use?

1. Set middleware


```go
    h := server.Default()
    h.Use(limiter.AdaptiveLimit())
```


2. Run example code


```go
import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/limiter"
)

func main() {
	h := server.Default(server.WithHostPorts(":1000"))
	h.Use(limiter.AdaptiveLimit())
	h.GET("/hello", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(consts.StatusOK, "hello")
	})
	h.Spin()
}
```
