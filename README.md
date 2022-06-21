# limiter (This is a community driven project)

##  Adaptivelimit

###  Adaptive Algorithm for [Hertz](https://github.com/cloudwego/hertz)


#### How to use?

1. Set middleware


```
    h := server.Default()
	h.Use(Adaptlivelimit())
```


2. Run example code


```
import (
	"context"


	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func main() {
	h := server.Default(server.WithHostPorts(":1000"))
	h.Use(Adaptlivelimit())
	h.GET("/hello", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(consts.StatusOK, "hello")
	})
	h.Spin()
}
```


Screenshot

More info
See example