# Adaptivelimit

Adaptive Algorithm for [Hertz](https://github.com/cloudwego/hertz)

## Core

```
    Statistics of /proc/state data, analysis of CPU load, and correction of CPU usage deviation using EMA algorithm
    
    cpu load :
        cpu = cpuᵗ⁻¹ * decay + cpuᵗ * (1 - decay)

    Flow limitation formula:
        (cpu.load > 80% || (now - predrop) < 1s) && MaxProcess < InFlight 
    1. cpu load greater than 80 %  and Over-limited flow occurs in one second, mainly to prevent flow fluctuations
    2. Is the maximum number of requests that can be passed in a sample period less than the current number of requests
  
```

## Usage

```
	h := server.Default(server.WithHostPorts(":1000"))
	h.Use(Adaptlivelimit())
	h.GET("/hello", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(consts.StatusOK, "hello")
	})
	h.Spin()
```



