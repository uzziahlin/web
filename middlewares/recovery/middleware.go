package recovery

import "github.com/uzziahlin/web"

type MiddlewareBuilder struct {
	StatusCode int
	ErrMsg     string
	LogFunc    func(ctx *web.Context)
}

func (b MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespStatus = b.StatusCode
					ctx.RespData = []byte(b.ErrMsg)
					b.LogFunc(ctx)
				}
			}()

			next(ctx)
		}
	}
}
