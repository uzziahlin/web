package web

type Middleware func(next HandleFunc) HandleFunc

func MiddlewareOptionBuilder(mdls ...Middleware) Option {
	return func(httpServer *DefaultHttpServer) {
		httpServer.mdls = mdls
	}
}
