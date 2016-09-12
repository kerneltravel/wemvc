package wemvc

import (
	"errors"
	"fmt"
	"strings"
	"runtime"
)

type routeTree struct {
	routeNode
	funcMap   map[string]RouteFunc
	MatchCase bool
}

func (tree *routeTree) addFunc(name string, fun RouteFunc) error {
	if len(name) == 0 {
		return errors.New("The parameter 'name' cannot be empty")
	}
	if fun == nil {
		return errors.New("The parameter 'fun' cannot be nil")
	}
	if _, ok := tree.funcMap[name]; ok {
		return fmt.Errorf("The '%s' function is already exist.", name)
	}
	tree.funcMap[name] = fun
	return nil
}

func (tree *routeTree) lookupDepth(indexNode *routeNode, pathLength uint16, urlParts []string, method string, endWithSlash bool) (found bool, ctrl *controllerInfo, routeMap map[string]string) {
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
		str1 := indexNode.Path
		str2 := curPath
		if !tree.MatchCase {
			str1 = strings.ToLower(str1)
			str2 = strings.ToLower(str2)
		}
		if str1 != str2 {
			return
		}
	} else if indexNode.NodeType == param {
		// deal with dynamic path
		dynPathSplits := indexNode.PathSplits // the dynamic route paths
		routeCheckIndex := 0
		for {
			// if the target path cursor and the dynamic route cursor both reach the end, it means the target path and the dynamic route path mach
			if len(curPath) == 0 && routeCheckIndex == len(dynPathSplits) {
				break
			}
			if (len(curPath) == 0 && routeCheckIndex != len(dynPathSplits)) || (len(curPath) != 0 && routeCheckIndex == len(dynPathSplits)) {
				return
			}
			dynPath := dynPathSplits[routeCheckIndex]
			if tree.isParamPath(dynPath) {
				paramName := dynPath[1:len(dynPath)-1]
				opt := indexNode.Params[paramName]
				checkFunc := tree.funcMap[opt.Validation]
				if checkFunc == nil {
					return
				}
				data := checkFunc(curPath, opt)
				if len(data) == 0 {
					return
				}
				routeData[paramName] = data
				curPath = curPath[len(data):]
			} else {
				str1 := curPath
				str2 := dynPath
				if !tree.MatchCase {
					str1 = strings.ToLower(str1)
					str2 = strings.ToLower(str2)
				}
				if strings.HasPrefix(str1, str2) {
					curPath = curPath[len(dynPath):]
				} else {
					return
				}
			}
			routeCheckIndex++
		}
	} else {
		return
	}
	if indexNode.CurDepth == pathLength {
		ctrl = indexNode.CtrlInfo
		routeMap = routeData
		// detect default value
		if ctrl == nil {
			f, c, rm := indexNode.detectDefault(method)
			if f {
				found = true
				ctrl = c
				if rm != nil {
					for key, value := range rm {
						routeMap[key] = value
					}
				}
			} else {
				found = false
				routeMap = nil
				ctrl = nil
			}
		} else {
			found = true
		}
		if found {
			a,ok := routeMap["action"]
			if ok && ! tree.validateAction(a, method, ctrl) {
				return false, nil, nil
			}
		}
		return
	}
	// check the last url parts
	if !indexNode.hasChildren() {
		return
	}
	for _, child := range indexNode.Children {
		ok, result, rd := tree.lookupDepth(child, pathLength, urlParts, method, endWithSlash)
		if ok {
			if rd != nil && len(rd) > 0 {
				if _,ok = rd["pathInfo"];ok {
					if a,ok := rd["action"];ok {
						if !tree.validateAction(a, method, result) {
							return false, nil, nil
						}
					}
				}
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
	return
}

func (tree *routeTree) lookup(urlPath, method string) (*controllerInfo, map[string]string, error) {
	if urlPath == "/" {
		ctrl := tree.CtrlInfo
		if ctrl == nil {
			f,c,r := tree.detectDefault(method)
			if f {
				return c, r, nil
			}
		} else {
			return nil, nil, nil
		}
		return tree.CtrlInfo, nil, nil
	}
	urlParts, err := splitURLPath(urlPath)
	if err != nil {
		return nil, nil, err
	}
	var pathLength = uint16(len(urlParts))
	if pathLength == 0 || len(tree.Children) == 0 {
		return nil, nil, nil
	}
	var endWithSlash = strings.HasSuffix(urlPath, "/")
	for _, child := range tree.Children {
		ok, result, rd := tree.lookupDepth(child, pathLength, urlParts, method, endWithSlash)
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
		}
		tree.CtrlInfo = ctrlInfo
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
	node.MatchCase = runtime.GOOS != "windows"
	return node
}
