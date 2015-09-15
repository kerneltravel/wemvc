package wemvc

import (
	"strings"
	"regexp"
	//"strconv"
)

type routeNode struct {
	pathStr    string
	depth      int
	controller IController

	parent   *routeNode
	children map[string]*routeNode
}

func (this *routeNode) appendChild(node *routeNode) {
	if node == nil || !this.pathNameValid(node.pathStr) {
		return
	}
	node.depth = this.depth + 1
	node.parent = this

	if this.children == nil {
		this.children = make(map[string]*routeNode)
	}
	var existingChild = this.children[node.pathStr]
	if existingChild == nil {
		this.children[node.pathStr] = node
	} else {
		// change the controller
		if existingChild.controller == nil && node.controller != nil {
			existingChild.controller = node.controller
		}
		// combine the child tree
		if node.children != nil {
			for _, child := range node.children {
				existingChild.appendChild(child)
			}
		}
	}
}

// the path regex validation
// 1. single string, such as "name", "1998", "Janey"
// 2. replacer, such as "{id}", "{id:int}"
// 3. single string and replacer, such as "name-{id}", "article_{year:int}"
func (this *routeNode)pathNameValid(p string) bool {
	if this.depth == 1 && p == "/" {
		return true
	}
	regex, _ := regexp.Compile(`^((\w+)|({\w+(:[\w]{0,})?}))+((_|-|\+| |)((\w+)|({\w+(:[\w]{0,})?})))*$`)
	return regex.MatchString(p)
}

func (this *routeNode) child(pathStr string, controller IController) *routeNode {
	//println("add child to", this.pathStr + "[" + strconv.Itoa(this.depth) + "]:", pathStr)
	var p string
	if this.depth == 0 && (pathStr == "/" || len(pathStr) == 0) {
		p = "home"
	} else {
		p = strings.Trim(pathStr, " ")
	}
	var node = &routeNode{
		pathStr:    p,
		depth:      this.depth + 1,
		controller: controller,
		parent:     this,
		children:   nil,
	}
	this.appendChild(node)
	return this.children[pathStr]
}

func (this *routeNode) matchPath(pathUrl string) (bool, IController) {
	if this.pathStr == pathUrl {
		return true, this.controller
	}
	return false, nil
}

func (this *routeNode) matchDepth(pathUrls []string) (bool, IController) {
	if len(pathUrls) == this.depth {
		return this.matchPath(pathUrls[len(pathUrls)-1])
	} else if len(pathUrls) > this.depth && this.children != nil {
		for _, child := range this.children {
			b, c := child.matchDepth(pathUrls)
			if b {
				return b, c
			}
		}
	}
	return false, nil
}

type routeTree struct {
	rootNode routeNode // the depth of the root node is 1
}

func (this *routeTree)AddController(p string, c IController) {
	if p == "/" {
		this.rootNode.controller = c
		return
	}
	var paths = strings.Split(strings.TrimSuffix(p, "/"), "/")
	var current = &(this.rootNode)
	for i := 1; i < len(paths);i++ {
		if i + 1 == len(paths) {
			current = current.child(paths[i], c)
		} else {
			current = current.child(paths[i], nil)
		}
	}
}