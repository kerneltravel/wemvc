package wemvc

import (
	"strings"
	"regexp"
	//"strconv"
	"errors"
)

type ruleReg struct {
	key string
	reg *regexp.Regexp
}

type ruleRegMap map[int]*ruleReg

func (this ruleRegMap)findByKey(key string) *ruleReg {
	for _, r := range this {
		if r.key == key {
			return r
		}
	}
	return nil
}

type routeNode struct {
	pathStr    string
	routeKeys  ruleRegMap
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
// 2. route key, such as "{id}", "{name}"
// 3. single string and replacer, such as "name-{id}", "article_{year}", "{year}+{month}_{day}.{hour}"
func (this *routeNode)pathNameValid(p string) bool {
	if this.depth == 1 && p == "/" {
		return true
	}
	regex, _ := regexp.Compile(`^((\w+)|({\w+}))+((_|-|\+|\.)((\w+)|({\w+})))*$`)
	return regex.MatchString(p)
}

func (this *routeNode) child(pathStr string, controller IController, rules ruleRegMap) *routeNode {
	//println("add child to", this.pathStr + "[" + strconv.Itoa(this.depth) + "]:", pathStr)
	var p = strings.Trim(pathStr, " ")

	var node = &routeNode{
		pathStr:    p,
		depth:      this.depth + 1,
		controller: controller,
		parent:     this,
		children:   nil,
		routeKeys:  nil,
	}

	var keys = regRouteKey.FindAllString(pathStr, -1)
	if len(keys) > 0 {
		var rc = make(ruleRegMap)
		for i := 0; i < len(keys); i++ {
			var r *ruleReg
			if rules != nil {
				r = rules.findByKey(keys[i])
			}
			if r == nil {
				r = &ruleReg{key:keys[i], reg:regString}
			}
			rc[i] = r
		}
		node.routeKeys = rc
	}

	this.appendChild(node)
	return this.children[pathStr]
}

func (this *routeNode) matchPath(pathUrl string, routeData map[string]string) (bool, IController) {
	if this.routeKeys != nil {
		if this.pathStr == pathUrl {
			return true, this.controller
		} else {
			return false, nil
		}
	}
	i,j,rd := 0,0,0 // i: the index of the this.pathStr j: the index of pathUrl
	for {
		if i == len(this.pathStr) && j == len(pathUrl) {
			return true, this.controller
		}
		if i == len(this.pathStr) || j == len(pathUrl) {
			return false, nil
		}
		r := string(this.pathStr[i:])
		v := string(pathUrl[j:])
		if strings.HasPrefix(r,"{") {
			finder,_ := regexp.Compile(`^{\w+}`)
			tmpKeys := finder.FindAllString(r, 1)
			if len(tmpKeys) < 1 {
				return false, nil
			}
			key := tmpKeys[0]
			if this.routeKeys[rd].key == key {
				values := this.routeKeys[rd].reg.FindAllString(v, 1)
				if len(values) < 1 {
					return false, nil
				}
				value := values[0]
				if len(values) < 1 {
					return false, nil
				}
				routeData[key] = value
				rd = rd + 1
				i = i + len(key)
				j = j + len(value)
			} else {
				return false, nil
			}
		} else {
			if this.pathStr[i] != pathUrl[j] {
				return false, nil
			}
			i = i + 1
			j = j + 1
		}
	}

	return false, nil
}

func (this *routeNode) matchDepth(pathUrls []string, routeData map[string]string) (bool, IController) {
	if this.depth > len(pathUrls) {
		return false, nil
	}
	var curPath = pathUrls[this.depth - 1]
	match, contr := this.matchPath(curPath, routeData)
	if !match {
		return false, nil
	}
	if len(pathUrls) == this.depth {
		return true, contr
	} else if len(pathUrls) > this.depth && this.children != nil {
		for _, child := range this.children {
			b, c := child.matchDepth(pathUrls, routeData)
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

func (this *routeTree)AddController(p string, c IController, valid ...string) {
	if p == "/" {
		this.rootNode.controller = c
		return
	}

	if !strings.HasPrefix(p, "/") {
		panic(errors.New("the route path should have prefix '/'"))
	}

	if strings.HasSuffix(p, "/") {
		panic(errors.New("the route path should not have suffix '/'"))
	}

	fixPath := strings.TrimSuffix(p, " ")
	if err:=this.checkRouteDataKey(fixPath);err!= nil{
		panic(err)
	}
	rules := this.genValidation(valid)

	var paths = strings.Split(fixPath, "/")
	var current = &(this.rootNode)
	for i := 1; i < len(paths);i++ {
		if i + 1 == len(paths) {
			current = current.child(paths[i], c, rules)
		} else {
			current = current.child(paths[i], nil, rules)
		}
	}
}

func (this *routeTree)checkRouteDataKey(paths string) error {
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

func (this *routeTree)genValidation(v []string) ruleRegMap {
	var result = make(ruleRegMap)
	finder,_ := regexp.Compile(`^{\w+}=`)
	index := 0
	for _, rule := range v {
		keys := finder.FindAllString(rule, 1)
		if len(keys) == 1 {
			key := strings.TrimSuffix(keys[0], "=")
			var reg *regexp.Regexp = nil
			regRule := strings.TrimLeft(rule, key + "=")
			if regRule == "num" || regRule == "number" {
				reg = regNumber
			} else if regRule == "string" {
				reg = regString
			} else {
				if !strings.HasPrefix(regRule, "^") {
					regRule = "^" + regRule
				}
				r,err := regexp.Compile(regRule)
				if err != nil {
					msg := "Failed to analyze the route key validation rule \"" + rule + "\". \r\n" +
					err.Error()
					panic(errors.New(msg))
				} else {
					reg = r
				}
			}
			if result.findByKey(key) != nil {
				panic(errors.New("Duplicate definition of the rule for the key \"" + key + "\""))
			}
			result[index] = &ruleReg{key:key, reg:reg}
		}
	}
	return result
}