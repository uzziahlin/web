package session

import (
	"github.com/uzziahlin/web"
)

type Manager struct {
	Store
	Propagator
	SessCtxKey string
}

func (m *Manager) GetSession(ctx *web.Context) (Session, error) {

	sess, ok := ctx.UserValues[m.SessCtxKey]

	if ok {
		return sess.(Session), nil
	}

	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any, 1)
	}

	sessID, err := m.Extract(ctx.Req)

	if err != nil {
		return nil, err
	}

	session, err := m.Get(ctx.Req.Context(), sessID)

	if err != nil {
		return nil, err
	}

	ctx.UserValues[m.SessCtxKey] = session

	return session, nil
}

// InitSession 初始化一个 session，并且注入到 http response 里面
func (m *Manager) InitSession(ctx *web.Context, id string) (Session, error) {
	sess, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	if err = m.Inject(id, ctx.Resp); err != nil {
		return nil, err
	}
	return sess, nil
}

// RefreshSession 刷新 Session
func (m *Manager) RefreshSession(ctx *web.Context) (Session, error) {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	// 刷新存储的过期时间
	err = m.Refresh(ctx.Req.Context(), sess.ID())
	if err != nil {
		return nil, err
	}
	// 重新注入 HTTP 里面
	if err = m.Inject(sess.ID(), ctx.Resp); err != nil {
		return nil, err
	}
	return sess, nil
}

// RemoveSession 删除 Session
func (m *Manager) RemoveSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	err = m.Store.Remove(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}
	return m.Propagator.Remove(ctx.Resp)
}
