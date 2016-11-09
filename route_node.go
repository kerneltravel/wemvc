package wemvc

import (
	"errors"
	"fmt"
	"strings"
)

type pathType uint8

const (
	rtStatic pathType = iota
	rtRoot
	rtParam
	rtCatchAll

	rtParamBegin = '<'
	rtParamBeginStr = "<"
	rtParamEnd = '>'
	rtParamEndStr = ">"
	rtPathInfo = "*pathInfo"
)

// RouteOption the route option struct
type RouteOption struct {
	Validation      string
	HasDefaultValue bool
	DefaultValue    string
	Setting         string
	MaxLength       uint8
	MinLength       uint8
}

type routeNode struct {
	NodeType   pathType
	CurDepth   uint16
	MaxDepth   uint16
	Path       string
	PathSplits []string
	Params     map[string]RouteOption
	CtrlInfo   *controllerInfo
	Children   []*routeNode
}

func (node *routeNode) isLeaf() bool {
	if node.NodeType == rtRoot {
		return false
	}
	return node.hasChildren() == false
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
		node.Children = append(node.Children, childNode)
		return nil
	}
	if childNode.MaxDepth > existChild.MaxDepth {
		existChild.MaxDepth = childNode.MaxDepth
	}
	if childNode.isLeaf() {
		if existChild.CtrlInfo != nil {
			return fmt.Errorf("Duplicate controller info in route tree. Path: %s, Depth: %d",
				existChild.Path,
				existChild.CurDepth)
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

func (node *routeNode) isParamPath(path string) bool {
	return strings.HasPrefix(path, rtParamBeginStr) && strings.HasSuffix(path, rtParamEndStr)
}

func (node *routeNode) detectDefault(method string) (bool, *controllerInfo, map[string]string) {
	if !node.hasChildren() {
		return false, nil, nil
	}
	for _, child := range node.Children {
		if child.NodeType != rtParam || len(child.PathSplits) != 1 || !node.isParamPath(child.PathSplits[0]) {
			continue
		}
		paramName := ""
		var opt RouteOption
		for name, o := range child.Params {
			paramName = name
			opt = o
			break
		}
		if !opt.HasDefaultValue {
			continue
		}
		if child.CtrlInfo != nil {
			if paramName == "action" {
				if ok, _ := child.CtrlInfo.containsAction(opt.DefaultValue, method); !ok {
					return false, nil, nil
				}
			}
			return true, child.CtrlInfo, map[string]string{paramName: opt.DefaultValue}
		}
		found, ctrl, routeMap := child.detectDefault(method)
		if found {
			if paramName == "action" {
				if ok, _ := ctrl.containsAction(opt.DefaultValue, method); !ok {
					return false, nil, nil
				}
			}
			routeMap[paramName] = opt.DefaultValue
			return true, ctrl, routeMap
		}
	}
	return false, nil, nil
}

func newRouteNode(routePath string, ctrlInfo *controllerInfo) (*routeNode, error) {
	err := checkRoutePath(routePath)
	if err != nil {
		return nil, err
	}
	splitPaths, err := splitURLPath(routePath)
	if err != nil {
		return nil, err
	}
	var length = uint16(len(splitPaths))
	if length == 0 {
		return nil, nil
	}
	if detectNodeType(splitPaths[length-1]) == rtCatchAll {
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
		if child.NodeType == rtParam {
			paramPath, params, err := analyzeParamOption(child.Path)
			if err != nil {
				return nil, err
			}
			child.PathSplits = paramPath
			child.Params = params
		}
		if result == nil {
			result = child
			current = result
		} else {
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
		if strings.Contains(current.Path, "*") && current.NodeType != rtCatchAll {
			return nil, errors.New("Invalid URL route parameter '" + current.Path + "'")
		}
		if current.NodeType == rtCatchAll && len(current.Children) > 0 {
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
