package wemvc

import (
	"errors"
	"strings"
	"fmt"
)

type routeNode struct {
	NodeType nodeType
	Depth    uint8
	Path     string
	Children []*routeNode
	CtrlInfo *controllerInfo

	parent   *routeNode
}

func (node *routeNode) isLeaf() bool {
	if node.NodeType == root {
		return false
	}
	if len(node.Children) == 0 {
		if node.CtrlInfo == nil {
			panic(errors.New("Invalid route node. The leaf of the route tree must have controller info"))
		} else {
			return true
		}
	}
	return false
}

func (node *routeNode) checkChild(path string) *routeNode {
	if len(node.Children) == 0 {
		return nil
	}
	for _, child := range node.Children {
		if child.Path == path {
			return child
		}
	}
	return nil
}

func (node *routeNode) addChild(childNode *routeNode) error {
	if childNode == nil {
		return errors.New("'childNode' parameter cannot be nil")
	}
	if len(node.Children) == 0 {
		childNode.parent = node
		node.Children = []*routeNode{childNode}
		return nil
	}
	var existChild = node.checkChild(childNode.Path)
	if existChild == nil {
		childNode.parent = node
		node.Children = append(node.Children, childNode)
		return nil
	}
	if childNode.isLeaf() {
		if existChild.CtrlInfo != nil {
			return errors.New(fmt.Sprintf("Duplicate controller info in route tree. Path: %s, Depth: %d", existChild.Path, existChild.Depth))
		}
		existChild.CtrlInfo = childNode.CtrlInfo
	} else {
		for _, child := range childNode.Children {
			err := existChild.addChild(child)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type rootNode struct {
	routeNode
}

func (root *rootNode) lookup(urlPath string, routeData *map[string]string) (*controllerInfo, error) {
	if urlPath == "/" {
		return root.CtrlInfo
	}
	urlParts,err := splitUrlPath(urlPath)
	if err != nil {
		return nil, err
	}
	if len(urlParts) == 0 {
		return nil,nil
	}

	return nil, nil
}

func (root *rootNode) addRoute(routePath string, ctrlInfo *controllerInfo) error {
	if len(routePath) == 0 {
		return errors.New("'routePath' param cannot be empty")
	}
	if ctrlInfo == nil {
		return errors.New("'ctrlInfo' param cannot be nil")
	}
	if routePath == "/" {
		if root.CtrlInfo != nil {
			return errors.New("Duplicate controller info for route '/'")
		} else {
			root.CtrlInfo = ctrlInfo
		}
		return nil
	}
	branch,err := newRouteBranch(routePath, ctrlInfo)
	if err != nil {
		return err
	}
	if err = root.addChild(branch); err != nil {
		return err
	}
	return nil
}

func newRootNode() *rootNode {
	var node = &rootNode{}
	node.NodeType = root
	node.Depth = 0
	node.Path = "/"
	node.CtrlInfo = nil
	return node
}

func splitUrlPath(urlPath string) ([]string, error) {
	if len(urlPath) == 0 {
		return nil, errors.New("The URL path is empty")
	}
	p := strings.Trim(urlPath, "/")
	splits := strings.Split(p, "/")
	var result []string
	for _, s := range splits {
		if len(s) == 0 || s == "." {
			continue
		}
		if s == ".." {
			return nil, errors.New("Invalid URL path. the URL path cannot contains '..'")
		}
		result = append(result, s)
	}
	return result, nil
}

func detectNodeType(p string) nodeType {
	if p == "/" {
		return root
	}
	if strings.Contains(p, "{") && strings.Contains(p, "}") {
		return param
	}
	if p == "*pathInfo" {
		return catchAll
	}
	return static
}

func newRouteBranch(routePath string, ctrlInfo *controllerInfo) (*routeNode, error) {
	splitPaths,err := splitUrlPath(routePath)
	if err != nil {
		return nil, err
	}
	if len(splitPaths) == 0 {
		return nil, nil
	}
	var result *routeNode
	var current *routeNode
	for i, p := range splitPaths {
		if result == nil {
			result = &routeNode{
				NodeType: detectNodeType(p),
				Depth: uint8(i+1),
				Path: p,
			}
			current = result
		} else {
			var child = &routeNode{
				NodeType: detectNodeType(p),
				Depth: uint8(i+1),
				Path: p,
			}
			current.Children = append(current.Children, child)
			current = current.Children[0]
		}
	}
	current.CtrlInfo = ctrlInfo
	current = result
	for{
		if strings.Contains(current.Path, "*") && current.Path != "*pathInfo" {
			return nil, errors.New("Invalid URL param '" + current.Path + "'")
		}
		if current.NodeType == catchAll && len(current.Children) > 0 {
			return nil, errors.New("*pathInfo")
		}
	}
	return result, nil
}