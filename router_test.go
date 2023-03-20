package web

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestTrieRouter_addRoute(t *testing.T) {

	mockHandler := func(*Context) {}

	testCases := []struct {
		method string
		path   string
	}{
		{
			method: "get",
			path:   "/",
		},
		{
			method: "get",
			path:   "/order/detail",
		},
		{
			method: "post",
			path:   "/user/info",
		},
		{
			method: "post",
			path:   "/order/detail/:id",
		},
		{
			method: "get",
			path:   "/user/*",
		},
		{
			method: "get",
			path:   "/user/*/detail",
		},
		{
			method: "get",
			path:   "/user/*/*/detail",
		},
		{
			method: "get",
			path:   "/user/*/charge",
		},
	}

	wantRouter := &trieRouter{
		trees: map[string]*node{
			"get": &node{
				path:    "/",
				handler: mockHandler,
				children: map[string]*node{
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path:    "detail",
								handler: mockHandler,
							},
						},
					},
					"user": &node{
						path: "user",
						starChild: &node{
							path:    "*",
							handler: mockHandler,
							children: map[string]*node{
								"detail": &node{
									path:    "detail",
									handler: mockHandler,
								},
								"charge": &node{
									path:    "charge",
									handler: mockHandler,
								},
							},
							starChild: &node{
								path: "*",
								children: map[string]*node{
									"detail": &node{
										path:    "detail",
										handler: mockHandler,
									},
								},
							},
						},
					},
				},
			},
			"post": &node{
				path: "/",
				children: map[string]*node{
					"user": &node{
						path: "user",
						children: map[string]*node{
							"info": &node{
								path:    "info",
								handler: mockHandler,
							},
						},
					},
					"order": &node{
						path: "order",
						children: map[string]*node{
							"detail": &node{
								path: "detail",
								paramChild: &node{
									path:    ":id",
									handler: mockHandler,
								},
							},
						},
					},
				},
			},
		},
	}

	testRouter := &trieRouter{}

	for _, tc := range testCases {
		testRouter.addRoute(tc.method, tc.path, mockHandler)
	}

	msg, ok := testRouter.equals(wantRouter)

	assert.True(t, ok, msg)

	testRouter1 := &trieRouter{}

	assert.Panicsf(t, func() {
		testRouter1.addRoute("get", "order/detail", mockHandler)
	}, "路径必须以[/]开头且不能以[/]结尾，请检查路由路径")

	assert.Panicsf(t, func() {
		testRouter1.addRoute("get", "/order/detail/", mockHandler)
	}, "路径必须以[/]开头且不能以[/]结尾，请检查路由路径")

	assert.Panicsf(t, func() {
		testRouter1.addRoute("get", "/order//detail", mockHandler)
	}, "路径必须以[/]开头且不能以[/]结尾，请检查路由路径")
}

func TestTrieRouter_matchRoute(t *testing.T) {

	mockHandler := func(*Context) {}

	testRoutes := []struct {
		method string
		path   string
	}{
		{
			method: "get",
			path:   "/",
		},
		{
			method: "get",
			path:   "/order/detail",
		},
		{
			method: "post",
			path:   "/user/info",
		},
		{
			method: "get",
			path:   "/user/*",
		},
		{
			method: "get",
			path:   "/user/:id(^[0-9]+$)",
		},
		{
			method: "get",
			path:   "/user/*/detail",
		},
		{
			method: "get",
			path:   "/user/*/charge",
		},
		{
			method: "post",
			path:   "/order/detail/:id",
		},
	}

	testRouter := &trieRouter{}

	for _, tc := range testRoutes {
		testRouter.addRoute(tc.method, tc.path, mockHandler)
	}

	testCases := []struct {
		name      string
		method    string
		path      string
		wantFound bool
		wantInfo  RouteInfo
	}{
		{
			name:      "get order detail",
			method:    "get",
			path:      "/order/detail",
			wantFound: true,
			wantInfo: RouteInfo{
				info: &node{
					path:    "detail",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "test common match",
			method:    "get",
			path:      "/user/pay/charge",
			wantFound: true,
			wantInfo: RouteInfo{
				info: &node{
					path:    "charge",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "test common match",
			method:    "get",
			path:      "/user/go/to/detail",
			wantFound: true,
			wantInfo: RouteInfo{
				info: &node{
					path:    "detail",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "test path variable",
			method:    "post",
			path:      "/order/detail/12",
			wantFound: true,
			wantInfo: RouteInfo{
				info: &node{
					path:    ":id",
					handler: mockHandler,
				},
				params: map[string]string{
					"id": "12",
				},
			},
		},
		{
			name:      "test multilevel",
			method:    "get",
			path:      "/user/a/b/c/d/e/detail",
			wantFound: true,
			wantInfo: RouteInfo{
				info: &node{
					path:    "detail",
					handler: mockHandler,
				},
			},
		},
		{
			name:      "test multilevel",
			method:    "get",
			path:      "/user/a/b/c/d/e",
			wantFound: true,
			wantInfo: RouteInfo{
				info: &node{
					path:    "*",
					handler: mockHandler,
					children: map[string]*node{
						"detail": &node{
							path:    "detail",
							handler: mockHandler,
						},
						"charge": &node{
							path:    "charge",
							handler: mockHandler,
						},
					},
				},
			},
		},
		{
			name:      "test regexp",
			method:    "get",
			path:      "/user/123456",
			wantFound: true,
			wantInfo: RouteInfo{
				info: &node{
					path:    ":id(^[0-9]+$)",
					handler: mockHandler,
				},
				params: map[string]string{
					"id": "123456",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found, ok := testRouter.matchRoute(tc.method, tc.path)

			assert.True(t, ok == tc.wantFound, "route not match")

			msg, ok := found.info.equals(tc.wantInfo.info)

			assert.True(t, ok, msg)

			assert.Equal(t, tc.wantInfo.params, found.params, "参数不相等")
		})
	}

}

func (r *trieRouter) equals(w *trieRouter) (string, bool) {

	for m, rn := range r.trees {
		wn, ok := w.trees[m]
		if !ok {
			return "no such method", false
		}
		msg, ok := rn.equals(wn)
		if !ok {
			return msg, false
		}
	}

	return "", true
}

func (n *node) equals(m *node) (string, bool) {
	if n.path != m.path {
		return "path is different", false
	}

	nHandler := reflect.ValueOf(n.handler)
	mHandler := reflect.ValueOf(m.handler)

	if nHandler != mHandler {
		return "handler is different", false
	}

	if len(n.children) != len(m.children) {
		return "count of child is different", false
	}

	for path, nChild := range n.children {
		mChild, ok := m.children[path]
		if !ok {
			return fmt.Sprintf("the path %s is not found in other node", path), false
		}

		msg, ok := nChild.equals(mChild)

		if !ok {
			return msg, false
		}
	}
	return "", true

}
