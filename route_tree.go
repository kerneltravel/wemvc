package wemvc

import (
	"errors"
	"fmt"
	"strings"
)

type routeTree struct {
	routeNode
	funcMap map[string]RouteFunc
}

func (tree *routeTree) addFunc(name string, fun RouteFunc) error {
	if tree.funcMap == nil {
		tree.funcMap = make(map[string]RouteFunc)
	}
	if len(name) == 0 {
		return errors.New("The parameter 'name' cannot be empty")
	}
	if fun == nil {
		return errors.New("The parameter 'fun' cannot be nil")
	}
	if _, ok := tree.funcMap[name]; ok {
		return errors.New(fmt.Sprintf("The '%s' function is already exist.", name))
	}
	tree.funcMap[name] = fun
	return nil
}

func (tree *routeTree) findFunc(name string) RouteFunc {
	if tree.funcMap == nil {
		return nil
	}
	return tree.funcMap[name]
}

func (tree *routeTree) lookupDepth(indexNode *routeNode, pathLength uint16, urlParts []string, endWithSlash bool) (found bool, ctrl *controllerInfo, routeMap map[string]string) {
	found = false
	ctrl = nil
	routeMap = nil
	if indexNode.MaxDepth+indexNode.CurDepth <= pathLength || indexNode.NodeType == root {
		return
	}
	var routeData = make(map[string]string)
	var curPath = urlParts[indexNode.CurDepth-1]
	if indexNode.NodeType == catchAll {
		// deal with *pathInfo
		var path string
		for _, part := range urlParts[indexNode.CurDepth-1:] {
			path = path + "/" + part
		}
		if endWithSlash {
			path = path + "/"
		}
		routeData["pathInfo"] = strings.TrimLeft(path, "/")
		found = true
		ctrl = indexNode.CtrlInfo
		routeMap = routeData
		return
	} else if indexNode.NodeType == static {
		// deal with static path
		var curPath = urlParts[indexNode.CurDepth-1]
		if indexNode.Path != curPath {
			return
		}
	} else if indexNode.NodeType == param {
		routePath := indexNode.ParamPath
		pathCheckIndex := 0
		routeCheckIndex := 0
		var paramNameBytes []byte
		isParamByte := false
		for {
			if pathCheckIndex >= len(curPath) && routeCheckIndex >= len(routePath) {
				break
			}
			if (pathCheckIndex >= len(curPath) && routeCheckIndex < len(routePath)) || (pathCheckIndex < len(curPath) && routeCheckIndex >= len(routePath)) {
				return
			}
			charOfRoute := routePath[routeCheckIndex]
			if charOfRoute != paramBegin && charOfRoute != paramEnd {
				if isParamByte {
					paramNameBytes = append(paramNameBytes, charOfRoute)
					routeCheckIndex++
					continue
				}
				if charOfRoute == curPath[pathCheckIndex] {
					pathCheckIndex++
					routeCheckIndex++
					continue
				} else {
					return
				}
			} else if charOfRoute == paramBegin {
				isParamByte = true
				routeCheckIndex++
				continue
			} else {
				paramName := string(paramNameBytes)
				paramNameBytes = nil
				isParamByte = false
				opt := indexNode.Params[paramName]
				checkFunc := tree.findFunc(opt.Validation)
				if checkFunc == nil {
					return
				}
				tempCurPath := curPath[pathCheckIndex:]
				data := checkFunc(tempCurPath, opt)
				if len(data) == 0 {
					return
				}
				routeData[paramName] = data
				routeCheckIndex++
				pathCheckIndex = pathCheckIndex + len(data)
			}
		}
	} else {
		return
	}
	if indexNode.CurDepth == pathLength {
		found = true
		ctrl = indexNode.CtrlInfo
		routeMap = routeData
		return
	}
	// check the last url parts
	if !indexNode.hasChildren() {
		return
	} else {
		for _, child := range indexNode.Children {
			ok, result, rd := tree.lookupDepth(child, pathLength, urlParts, endWithSlash)
			if ok {
				if rd != nil && len(rd) > 0 {
					for key, value := range rd {
						routeData[key] = value
					}
				}
				found = true
				ctrl = result
				routeMap = routeData
				return
			}
		}
	}
	return
}

func (tree *routeTree) lookup(urlPath string) (*controllerInfo, map[string]string, error) {
	if urlPath == "/" {
		return tree.CtrlInfo, nil, nil
	}
	urlParts, err := splitUrlPath(urlPath)
	if err != nil {
		return nil, nil, err
	}
	var pathLength = uint16(len(urlParts))
	if pathLength == 0 || len(tree.Children) == 0 {
		return nil, nil, nil
	}
	var endWithSlash = strings.HasSuffix(urlPath, "/")
	for _, child := range tree.Children {
		ok, result, rd := tree.lookupDepth(child, pathLength, urlParts, endWithSlash)
		if ok {
			return result, rd, nil
		}
	}
	return nil, nil, nil
}

func (tree *routeTree) addRoute(routePath string, ctrlInfo *controllerInfo) error {
	if len(routePath) == 0 {
		return errors.New("'routePath' param cannot be empty")
	}
	if ctrlInfo == nil {
		return errors.New("'ctrlInfo' param cannot be nil")
	}
	if routePath == "/" {
		if tree.CtrlInfo != nil {
			return errors.New("Duplicate controller info for route '/'")
		} else {
			tree.CtrlInfo = ctrlInfo
		}
		return nil
	}
	branch, err := newRouteNode(routePath, ctrlInfo)
	if err != nil {
		return err
	}
	if err = tree.addChild(branch); err != nil {
		return err
	}
	return nil
}

func newRouteTree() *routeTree {
	var node = &routeTree{
		funcMap: map[string]RouteFunc{
			"int":    intCheck,
			"string": stringCheck,
			"word":   wordCheck,
			"enum":   enumCheck,
		},
	}
	node.NodeType = root
	node.CurDepth = 0
	node.MaxDepth = 0
	node.Path = "/"
	node.CtrlInfo = nil
	return node
}
