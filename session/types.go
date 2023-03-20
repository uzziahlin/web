package session

import (
	"context"
	"net/http"
)

type Session interface {
	Set(ctx context.Context, key string, value any) error

	Get(ctx context.Context, key string) (any, error)

	ID() string
}

type Store interface {
	Generate(ctx context.Context, id string) (Session, error)

	Refresh(ctx context.Context, id string) error

	Remove(ctx context.Context, id string) error

	Get(ctx context.Context, id string) (Session, error)
}

type Propagator interface {
	// Inject 将 session id 注入到里面
	// Inject 必须是幂等的
	Inject(id string, writer http.ResponseWriter) error
	// Extract 将 session id 从 http.Request 中提取出来
	// 例如从 cookie 中将 session id 提取出来
	Extract(req *http.Request) (string, error)

	// Remove 将 session id 从 http.ResponseWriter 中删除
	// 例如删除对应的 cookie
	Remove(writer http.ResponseWriter) error
}
