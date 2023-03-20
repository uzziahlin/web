package accesslog

import (
	"encoding/json"
	"fmt"
	"github.com/uzziahlin/web"
)

type MiddlewareBuilder struct {
	log func(data string)
}

func NewMiddlewareBuilder() *MiddlewareBuilder {
	return &MiddlewareBuilder{
		log: func(data string) {
			fmt.Println(data)
		},
	}
}

func (b *MiddlewareBuilder) LogFunc(log func(string)) *MiddlewareBuilder {
	b.log = log
	return b
}

func (b *MiddlewareBuilder) Build() web.Middleware {
	return func(next web.HandleFunc) web.HandleFunc {
		return func(ctx *web.Context) {

			defer func() {
				al := accessLog{
					Path:         ctx.Req.URL.Path,
					HttpMethod:   ctx.Req.Method,
					Host:         ctx.Req.Host,
					MatchedRoute: ctx.MatchedRoute,
				}

				data, _ := json.Marshal(al)

				b.log(string(data))

			}()

			next(ctx)

		}
	}
}

type accessLog struct {
	Host         string
	HttpMethod   string
	Path         string
	MatchedRoute string
}
