package web

import (
	"regexp"
	"strings"
	"sync"
)

type HandleFunc func(*Context)

type RouteInfo struct {
	info   *node
	params map[string]string

	mdls []Middleware
}

func (rf *RouteInfo) addValue(key, value string) {
	if rf.params == nil {
		rf.params = make(map[string]string)
	}
	rf.params[key] = value
}

var compile = regexp.MustCompile("^:(.+)\\((.+)\\)$")

var _ iRouter = &trieRouter{}

// 路由抽象，定义路由的基本操作
type iRouter interface {
	// 注册路由
	addRoute(string, string, HandleFunc, ...Middleware)

	matchRoute(string, string) (RouteInfo, bool)
}

// 路由树节点
type node struct {
	path      string
	children  map[string]*node
	starChild *node

	paramChild *node
	paramName  string

	regexChild *node
	regexp     string

	handler HandleFunc

	route string

	mdls []Middleware

	cacheMdls []Middleware
}

// getOrCreateChild 判断当前节点是否存在path为参数的子节点
//	存在则返回，不存在则创建
func (n *node) getOrCreateChild(path string) *node {

	//判断是否是路径参数
	if path[0] == ':' {
		param, exp, ok := isRegexp(path)
		if ok {
			rChild := n.regexChild
			if rChild == nil {
				rChild = &node{
					path:      path,
					paramName: param,
					regexp:    exp,
				}
				n.regexChild = rChild
			} else if rChild.path != path {
				panic("不能注册不同的正则匹配路径")
			}
			return rChild
		}

		pChild := n.paramChild
		if pChild == nil {
			pChild = &node{
				path:      path,
				paramName: path[1:],
			}
			n.paramChild = pChild
		} else if pChild.path != path {
			panic("不能注册不同的路径参数")
		}
		return pChild
	}

	//判断是否是通配符
	if path == "*" {
		sChild := n.starChild
		if sChild == nil {
			sChild = &node{
				path: path,
			}
			n.starChild = sChild
		}
		return sChild
	}

	if n.children == nil {
		n.children = map[string]*node{}
	}

	child, ok := n.children[path]
	if !ok {
		child = &node{
			path: path,
		}
		n.children[path] = child
	}
	return child
}

// isRegexp 用来判断路径是否是正则匹配路径
// 如果是符合正则规则，则第一个参数返回参数名，第二个参数返回正则表达式
// 第三个参数返回是否匹配成功
func isRegexp(path string) (string, string, bool) {
	subMatch := compile.FindAllStringSubmatch(path, -1)
	if subMatch == nil {
		return "", "", false
	}
	// 验证正则表达式是否合法
	_, err := regexp.Compile(subMatch[0][2])

	if err != nil {
		panic("正则匹配路径，正则表达式不合法")
	}

	return subMatch[0][1], subMatch[0][2], true
}

// childOf 获取当前节点path为参数的子节点
// 第一个返回值为获得的节点
// 第二个返回值为是否是路径参数或者正则匹配参数，便于上游作特殊处理，
// 第三个返回值表示是否获取到节点
func (n *node) childOf(path string) (*node, bool, bool) {
	// 静态匹配优先级最高
	child, ok := n.children[path]
	if !ok {

		// 优先进行正则匹配
		if n.regexChild != nil {
			rChild := n.regexChild
			match := n.isMatch(path)
			if match {
				return rChild, true, true
			}
		}

		// 匹配路径参数
		if n.paramChild != nil {
			return n.paramChild, true, true
		}

		// 通配符匹配
		if n.starChild != nil {
			return n.starChild, false, true
		}

		// 通配符贪心, 使通配符可以匹配多级路径
		if n.path == "*" {
			return n, false, true
		}
	}
	return child, false, ok
}

// isMatch 判断路径是否满足正则匹配
// 是否有必要设计成node的行为？？？
func (n *node) isMatch(path string) bool {
	matched, err := regexp.Match(n.regexp, []byte(path))

	if err != nil {
		return false
	}

	return matched
}

// 前缀树路由实现
type trieRouter struct {
	trees map[string]*node

	mu sync.Mutex
}

// addRoute 提供路由注册功能
// method http请求方式
// path http请求路径
// handler 用户业务处理逻辑
func (r *trieRouter) addRoute(method, path string, handler HandleFunc, mdls ...Middleware) {

	if r.trees == nil {
		r.trees = map[string]*node{}
	}

	root := r.trees[method]
	if root == nil {
		root = &node{
			path: "/",
		}
		r.trees[method] = root
	}

	if path == "/" {
		root.handler = handler
		return
	}

	if path[0] != '/' || path[len(path)-1] == '/' {
		panic("路径必须以[/]开头且不能以[/]结尾，请检查路由路径")
	}

	// 切分路径，按 '/' 进行切分
	paths := strings.Split(path[1:], "/")

	// 将路径添加到前缀树
	for _, p := range paths {

		if p == "" {
			panic("不能有连续的[/], 请检查路由")
		}

		child := root.getOrCreateChild(p)
		root = child
	}

	root.route = path

	if handler != nil {
		root.handler = handler
	}

	if mdls != nil {
		root.mdls = append(root.mdls, mdls...)
	}

}

// matchRoute 路由匹配
func (r *trieRouter) matchRoute(method, path string) (RouteInfo, bool) {

	result := RouteInfo{}

	if r.trees == nil {
		return result, false
	}

	root, ok := r.trees[method]

	if !ok {
		return result, false
	}

	result.mdls = root.mdls

	if path == "/" {
		result.info = root
		return result, true
	}

	paths := strings.Split(path[1:], "/")

	// mdlsC := root.findMiddlewares(paths, Conditions...)

	cur := root

	for _, p := range paths {
		child, isParam, ok := cur.childOf(p)
		if !ok {
			return result, false
		}
		if isParam {
			result.addValue(child.paramName, p)
		}
		cur = child
	}

	result.info = cur

	if cur.cacheMdls == nil {
		mdlsC := root.findMiddlewares(paths, Conditions...)
		cur.cacheMdls = <-mdlsC
	}

	result.mdls = cur.cacheMdls

	return result, true
}

type qElem struct {
	level int
	elem  *node
}

func newElem(level int, elem *node) qElem {
	return qElem{
		level: level,
		elem:  elem,
	}
}

// findMiddlewares 广度优先遍历异步搜索Middleware
func (n *node) findMiddlewares(paths []string, conds ...Condition) <-chan []Middleware {
	c := make(chan []Middleware)

	go func() {

		var result Queue[Middleware]

		/*queue := []qElem{
			{
				level: 0,
				elem:  root,
			},
		}*/

		queue := Queue[qElem]{
			{
				level: 0,
				elem:  n,
			},
		}

		for queue.Len() > 0 {
			elem := queue.Pop()
			curNode := elem.elem
			path := paths[elem.level]

			for _, cond := range conds {

				if n, ok := cond(curNode, path); ok {
					result.Push(n.mdls...)

					if elem.level < len(paths)-1 {
						queue.Push(newElem(elem.level+1, n))
					}

				}

			}

			//// 判断是否有完全匹配的子节点，如果有则将子节点的middlewares拿出来，并且将子节点放到队列中
			//if n, ok := curNode.children[path]; ok {
			//	result.Push(n.mdls...)
			//
			//	queue.Push(newElem(elem.level+1, n))
			//
			//}
			//
			//// 判断是否存在路径参数，如果存在则将子节点的middlewares拿出来，并且将子节点放到队列中
			//if pn := curNode.paramChild; pn != nil {
			//	result.Push(pn.mdls...)
			//
			//	queue.Push(newElem(elem.level+1, pn))
			//}
			//
			//// 判断是否存在正则匹配，如果存在则将子节点的middlewares拿出来，并且将子节点放到队列中
			//if rn := curNode.regexChild; rn != nil && rn.isMatch(path) {
			//	result.Push(rn.mdls...)
			//
			//	queue.Push(newElem(elem.level+1, rn))
			//}
			//
			//// 判断子节点是否存在通配符，如果存在，则将middlewares拿出来，并且放到队列中
			//if sn := curNode.starChild; sn != nil {
			//	result.Push(sn.mdls...)
			//
			//	queue.Push(newElem(elem.level+1, sn))
			//
			//}

		}

		c <- result

	}()

	return c
}
