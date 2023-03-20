package web

import (
	"net"
	"net/http"
)

var _ HttpServer = &DefaultHttpServer{}

// Server 抽象 管理server的生命周期信息以及路由注册操作
type Server interface {
	Start() error
	iRouter
	// Use 提供插件注册功能
	Use(method, path string, mdls ...Middleware)
}

// HttpServer 对 Httpserver 进行抽象，管理http相关的操作
type HttpServer interface {
	Server
	http.Handler
	Get(string, HandleFunc)
	Post(string, HandleFunc)
}

// DefaultHttpServer 默认实现
type DefaultHttpServer struct {
	addr string
	iRouter

	mdls []Middleware

	t TemplateEngine
}

type Option func(httpServer *DefaultHttpServer)

func NewHttpServer(addr string, opts ...Option) HttpServer {
	server := &DefaultHttpServer{
		addr:    addr,
		iRouter: &trieRouter{},
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func (s *DefaultHttpServer) Get(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodGet, path, handleFunc)
}

func (s *DefaultHttpServer) Post(path string, handleFunc HandleFunc) {
	s.addRoute(http.MethodPost, path, handleFunc)
}

func (s *DefaultHttpServer) Use(method, path string, mdls ...Middleware) {
	s.addRoute(method, path, nil, mdls...)
}

// Start 启动Server
func (s *DefaultHttpServer) Start() error {
	// 监听端口
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	// 中间可以有生命周期回调

	return http.Serve(l, s)

}

// ServeHTTP 作为请求入口，处理Http请求
func (s *DefaultHttpServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// todo 路由匹配，封装上下文，调用业务处理逻辑
	ctx := &Context{
		Req:  req,
		Resp: resp,
		T:    s.t,
	}

	root := s.Serve

	// 拼接责任链
	for i := len(s.mdls) - 1; i >= 0; i-- {
		root = s.mdls[i](root)
	}

	// 处理输出数据
	root = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			defer func() {
				ctx.Resp.WriteHeader(ctx.RespStatus)
				_, _ = ctx.Resp.Write(ctx.RespData)
			}()
			next(ctx)
		}
	}(root)

	root(ctx)
}

func (s *DefaultHttpServer) Serve(ctx *Context) {
	route, ok := s.matchRoute(ctx.Req.Method, ctx.Req.URL.Path)

	if !ok || route.info.handler == nil {
		ctx.Resp.WriteHeader(http.StatusNotFound)
		_, _ = ctx.Resp.Write([]byte("resource not found"))
		return
	}

	ctx.PathParams = route.params
	ctx.MatchedRoute = route.info.route

	root := route.info.handler

	mdls := route.mdls

	for i := len(mdls) - 1; i >= 0; i-- {
		root = mdls[i](root)
	}

	root(ctx)

	// route.info.handler(ctx)
}
