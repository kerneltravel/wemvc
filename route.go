package wemvc

import (
	"errors"
	"regexp"
	"strings"
)

type routeNode struct {
	pathStr   string
	routeKeys map[string]*regexp.Regexp
	depth     int
	cInfo     *controllerInfo
	parent    *routeNode
	children  map[string]*routeNode
}

// append child of the route tree
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
		if existingChild.cInfo == nil && node.cInfo != nil {
			existingChild.cInfo = node.cInfo
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
// 2. route key, such as "{id}", "{name}"
// 3. single string and replacer, such as "name-{id}", "article_{year}", "{year}+{month}_{day}.{hour}"
func (this *routeNode) pathNameValid(p string) bool {
	if this.depth == 1 && p == "/" {
		return true
	}
	regex, _ := regexp.Compile(`^((\w+)|({\w+}))+((_|-|\+|\.)((\w+)|({\w+})))*$`)
	return regex.MatchString(p)
}

// build child of the route tree
func (this *routeNode) buildChild(pathStr string, cInfo *controllerInfo, rules map[string]*regexp.Regexp) *routeNode {
	var p = strings.Trim(pathStr, " ")

	var node = &routeNode{
		pathStr:   p,
		depth:     this.depth + 1,
		cInfo:     cInfo,
		parent:    this,
		children:  nil,
		routeKeys: nil,
	}

	var keys = regRouteKey.FindAllString(pathStr, -1)
	if len(keys) > 0 {
		var rc = make(map[string]*regexp.Regexp)
		for i := 0; i < len(keys); i++ {
			var r *regexp.Regexp
			if rules != nil {
				r = rules[keys[i]]
			}
			if r == nil {
				r = regString
			}
			rc[keys[i]] = r
		}
		node.routeKeys = rc
	}

	this.appendChild(node)
	return this.children[pathStr]
}

func (this *routeNode) matchPath(pathUrl string) (bool, *controllerInfo, map[string]string) {
	routeData := make(map[string]string)
	if this.routeKeys == nil {
		if this.pathStr == pathUrl {
			return true, this.cInfo, routeData
		} else {
			return false, nil, nil
		}
	}
	i, j := 0, 0 // i: the index of the this.pathStr j: the index of pathUrl
	for {
		if i == len(this.pathStr) && j == len(pathUrl) {
			return true, this.cInfo, routeData
		}
		if i == len(this.pathStr) || j == len(pathUrl) {
			return false, nil, nil
		}
		r := string(this.pathStr[i:])
		v := string(pathUrl[j:])
		if strings.HasPrefix(r, "{") {
			finder, _ := regexp.Compile(`^{\w+}`)
			tmpKeys := finder.FindAllString(r, 1)
			if len(tmpKeys) < 1 {
				return false, nil, nil
			}
			key := tmpKeys[0]
			reg := this.routeKeys[key]
			if reg != nil {
				values := reg.FindAllString(v, 1)
				if len(values) < 1 {
					return false, nil, nil
				}
				value := values[0]
				if len(values) < 1 {
					return false, nil, nil
				}
				routeData[key] = value
				i = i + len(key)
				j = j + len(value)
			} else {
				return false, nil, nil
			}
		} else {
			if this.pathStr[i] != pathUrl[j] {
				return false, nil, nil
			}
			i = i + 1
			j = j + 1
		}
	}
	return false, nil, nil
}

func (this *routeNode) matchDepth(method string, pathUrls []string, routeData map[string]string) (match bool, cType *controllerInfo, action string) {
	match = false
	cType = nil
	action = ""
	if this.depth > len(pathUrls) {
		return false, nil, ""
	}
	var curPath = pathUrls[this.depth-1]
	match, cType, r := this.matchPath(curPath)
	if !match {
		return
	} else {
		if routeData == nil {
			routeData = make(map[string]string)
		}
		for key, value := range r {
			if key == "{action}" {
				action = strings.ToLower(method + value)
			}
			routeData[key] = value
		}
	}
	if len(pathUrls) == this.depth {
		if len(action) < 1 {
			action = strings.ToLower(method)
		}
		match = cType.containsAction(action)
		return
	} else if len(pathUrls) > this.depth && this.children != nil {
		for _, child := range this.children {
			match, cType, action = child.matchDepth(method, pathUrls, routeData)
			if match {
				if len(action) < 1 {
					action = strings.ToLower(action)
				}
				match = cType.containsAction(action)
				return
			}
		}
	}
	return
}

type routeTree struct {
	rootNode routeNode // the depth of the root node is 1
}

func (this *routeTree) AddController(p string, cInfo *controllerInfo, valid ...string) {
	if p == "/" {
		this.rootNode.cInfo = cInfo
		return
	}

	if !strings.HasPrefix(p, "/") {
		panic(errors.New("the route path should has prefix '/'."))
	}

	if strings.HasSuffix(p, "/") {
		panic(errors.New("the route path should not has suffix '/'."))
	}

	fixPath := strings.TrimSuffix(p, " ")
	if err := this.checkRouteDataKey(fixPath); err != nil {
		panic(err)
	}
	rules := this.genValidation(valid)

	var paths = strings.Split(fixPath, "/")
	var current = &(this.rootNode)
	for i := 1; i < len(paths); i++ {
		if i+1 == len(paths) {
			current = current.buildChild(paths[i], cInfo, rules)
		} else {
			current = current.buildChild(paths[i], nil, rules)
		}
	}
}

func (this *routeTree) checkRouteDataKey(paths string) error {
	var keys = regRouteKey.FindAllString(paths, -1)
	if len(keys) < 1 {
		return nil
	}
	var s = ""
	for _, k := range keys {
		if strings.Contains(s, k) {
			return errors.New("Failed to add the route \"" + paths +
				"\". The route key \"" + k + "\" must be unique.")
		} else {
			s = s + k
		}
	}
	return nil
}

func (this *routeTree) genValidation(v []string) map[string]*regexp.Regexp {
	var result = make(map[string]*regexp.Regexp)
	finder, _ := regexp.Compile(`^{\w+}=`)
	for _, rule := range v {
		keys := finder.FindAllString(rule, 1)
		if len(keys) == 1 {
			key := strings.TrimSuffix(keys[0], "=")
			var reg *regexp.Regexp = nil
			regRule := strings.TrimLeft(rule, key+"=")
			if regRule == "num" || regRule == "number" {
				reg = regNumber
			} else if regRule == "string" {
				reg = regString
			} else {
				if !strings.HasPrefix(regRule, "^") {
					regRule = "^" + regRule
				}
				r, err := regexp.Compile(regRule)
				if err != nil {
					msg := "Failed to analyze the route key validation rule \"" + rule + "\". \r\n" + err.Error()
					panic(errors.New(msg))
				} else {
					reg = r
				}
			}
			if result[key] != nil {
				panic(errors.New("Duplicate definition of the rule for the key \"" + key + "\""))
			}
			result[key] = reg
		}
	}
	return result
}
