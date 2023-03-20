# Web框架
一个基于go语言编写的简单web框架，支持路由、中间件、模板、静态文件、session等功能。用户也可以很好地进行扩展路由树等核心功能。

使用示例如下：
```go
package main

import (
    "github.com/uzziahlin/web"
)

func main() {
    server := web.NewHttpServer(":8080")
    
    server.Get("/", func(ctx *web.Context) {
        ctx.WriteJSON(200, "Hello World!")
    })
    
    server.Get("/user/:id", func(ctx *web.Context) {
        ctx.WriteJSONOK("Hello User!")
    })
    
    server.Post("/user", func(ctx *web.Context) {
        ctx.WriteJSON(200, "Hello Post!")
    })
    
    server.Start()
}
```

