package web

type Condition func(*node, string) (*node, bool)

func FullMatchCond(n *node, path string) (*node, bool) {
	nn, ok := n.children[path]
	return nn, ok
}

func StarMatchCond(n *node, path string) (*node, bool) {
	return n.starChild, n.starChild != nil
}

func ParamMatchCond(n *node, path string) (*node, bool) {
	return n.paramChild, n.paramChild != nil
}

func RegexpMatchCond(n *node, path string) (*node, bool) {
	rn := n.regexChild
	return rn, rn != nil && rn.isMatch(path)
}

var Conditions = []Condition{
	FullMatchCond,
	ParamMatchCond,
	RegexpMatchCond,
	StarMatchCond,
}
