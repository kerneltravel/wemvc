package wemvc

import (
	"errors"
	"fmt"
	"strings"
)

type routeNode struct {
	NodeType nodeType
	CurDepth uint16
	MaxDepth uint16
	Path     string
	CtrlInfo *controllerInfo
	Children []*routeNode

	parent *routeNode
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

func max(a,b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

func (node *routeNode) hasChildren() bool {
	return len(node.Children) > 0
}

func (node *routeNode) findChild(path string) *routeNode {
	if !node.hasChildren() {
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
	var existChild = node.findChild(childNode.Path)
	if existChild == nil {
		childNode.parent = node
		node.Children = append(node.Children, childNode)
		return nil
	}
	existChild.MaxDepth = max(existChild.MaxDepth, childNode.MaxDepth)
	if childNode.isLeaf() {
		if existChild.CtrlInfo != nil {
			return errors.New(fmt.Sprintf("Duplicate controller info in route tree. Path: %s, Depth: %d", existChild.Path, existChild.CurDepth))
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

func (root *rootNode) lookupDepth(indexNode *routeNode, pathLength uint16, urlParts []string, endWithSlash bool) (bool, *controllerInfo, map[string]string) {
	if indexNode.MaxDepth + indexNode.CurDepth <= pathLength {
		return false, nil, nil
	}
	// TODO: deal with *pathInfo
	if indexNode.Path == "*pathInfo" {
		var path string
		for _, part := range urlParts[indexNode.CurDepth - 1:] {
			path = path + "/" + part
		}
		if endWithSlash {
			path = path + "/"
		}
		return true, indexNode.CtrlInfo, map[string]string {
			"pathInfo": path,
		}
	}
	// TODO: check path code and fill the route data
	var routeData = make(map[string]string)
	var curPath = urlParts[indexNode.CurDepth - 1]
	if indexNode.Path != curPath {
		return false, nil, nil
	}
	if indexNode.CurDepth == pathLength {
		return true, indexNode.CtrlInfo, routeData
	}
	// check the last url parts
	if !indexNode.hasChildren() {
		return false, nil, nil
	} else {
		for _, child := range indexNode.Children {
			ok, result, rd := root.lookupDepth(child, pathLength, urlParts, endWithSlash)
			if ok {
				if rd != nil && len(rd) > 0 {
					for key, value := range rd {
						routeData[key] = value
					}
				}
				return true, result, routeData
			}
		}
	}
	return false, nil, nil
}

func (root *rootNode) lookup(urlPath string) (*controllerInfo, map[string]string, error) {
	if urlPath == "/" {
		return root.CtrlInfo, nil, nil
	}
	urlParts, err := splitUrlPath(urlPath)
	if err != nil {
		return nil, nil, err
	}
	var pathLength = uint16(len(urlParts))
	if pathLength == 0 || len(root.Children) == 0 {
		return nil, nil, nil
	}
	var endWithSlash = strings.HasSuffix(urlPath, "/")
	for _, child := range root.Children {
		ok,result,rd := root.lookupDepth(child, pathLength, urlParts, endWithSlash)
		if ok {
			return result, rd, nil
		}
	}
	return nil, nil, nil
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
	branch, err := newRouteBranch(routePath, ctrlInfo)
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
	node.CurDepth = 0
	node.MaxDepth = 0
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
			return nil, errors.New("Invalid URL path. The URL path cannot contains '..'")
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
	splitPaths, err := splitUrlPath(routePath)
	if err != nil {
		return nil, err
	}
	var length = uint16(len(splitPaths))
	if length == 0 {
		return nil, nil
	}
	if detectNodeType(splitPaths[length - 1]) == catchAll {
		length = 255
	}
	var result *routeNode
	var current *routeNode
	for i, p := range splitPaths {
		var child = &routeNode{
			NodeType: detectNodeType(p),
			CurDepth: uint16(i + 1),
			MaxDepth: uint16(length - uint16(i)),
			Path:     p,
		}
		if result == nil {
			result = child
			current = result
		} else {
			child.parent = current
			current.Children = []*routeNode{child}
			current = current.Children[0]
		}
	}
	current.CtrlInfo = ctrlInfo
	current = result
	for {
		if current == nil {
			break
		}
		if strings.Contains(current.Path, "*") && current.NodeType != catchAll {
			return nil, errors.New("Invalid URL route parameter '" + current.Path + "'")
		}
		if current.NodeType == catchAll && len(current.Children) > 0 {
			return nil, errors.New("Invalid route'" + routePath + ". " +
				"The '*pathInfo' parameter should be at the end of the route. " +
				"For example: '/shell/*pathInfo'.")
		}
		if len(current.Children) > 0 {
			current = current.Children[0]
		} else {
			current = nil
		}
	}
	return result, nil
}
