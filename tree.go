// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package wemvc

import (
	"strings"
	"unicode"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func countParams(path string) uint8 {
	var n uint
	for i := 0; i < len(path); i++ {
		if path[i] != ':' && path[i] != '*' {
			continue
		}
		n++
	}
	if n >= 255 {
		return 255
	}
	return uint8(n)
}
/*
func checkPathChar(c byte, withNumber bool) bool {
	if withNumber {
		return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_'
	} else {
		return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	}
}

func countParams2(path string) ([]string, error) {
	var routeParams []string

	var paramChars []byte
	var inParamChar = false

	for i := 0; i < len(path); i++ {
		if path[i] == '{' {
			if len(paramChars) == 0 {
				inParamChar = true
				continue
			} else {
				return nil, errors.New(fmt.Sprintf("the route param has no closing character '}': %d", i))
			}
		}
		if path[i] == '}' {
			// check and ensure current route param is not empty
			if len(paramChars) == 0 {
				return nil, errors.New(fmt.Sprintf("the route param cannot be empty like \"{}\": %d", i))
			}
			var curParam = string(paramChars)
			for _, tmp := range routeParams {
				if tmp == curParam {
					return nil, errors.New(fmt.Sprintf("Duplicate route param \"%s\": %d", curParam, i))
				}
			}
			routeParams = append(routeParams, curParam)
			paramChars = make([]byte,0)
			inParamChar = false
			continue
		}
		if inParamChar {
			if len(paramChars) == 0 {
				if checkPathChar(path[i], false) {
					paramChars = append(paramChars, path[i])
				} else {
					return nil, errors.New(fmt.Sprintf("Invalid character '%c' at the beginin of the route param: %d", path[i], i))
				}
			} else {
				if checkPathChar(path[i], true) {
					paramChars = append(paramChars, path[i])
				} else {
					return nil, errors.New(fmt.Sprintf("Invalid character '%c' at the beginin of the route param: %d", path[i], i))
				}
			}
		}
	}
	if len(routeParams) > 255 {
		return nil, errors.New("Too many route params: the maximum number of the route param is 255")
	}
	return routeParams, nil
}
*/
type nodeType uint8

const (
	static nodeType = iota // default
	root
	param
	catchAll
)

type node struct {
	Path      string
	WildChild bool
	NodeType  nodeType
	MaxParams uint8
	Indices   string
	Children  []*node
	CtrlInfo  *controllerInfo
	Priority  uint32
}

// increments priority of the given child and reorders if necessary
func (n *node) incrementChildPrio(pos int) int {
	n.Children[pos].Priority++
	prio := n.Children[pos].Priority

	// adjust position (move to front)
	newPos := pos
	for newPos > 0 && n.Children[newPos-1].Priority < prio {
		// swap node positions
		tmpN := n.Children[newPos-1]
		n.Children[newPos-1] = n.Children[newPos]
		n.Children[newPos] = tmpN

		newPos--
	}

	// build new index char string
	if newPos != pos {
		n.Indices = n.Indices[:newPos] + // unchanged prefix, might be empty
			n.Indices[pos:pos+1] + // the index char we move
			n.Indices[newPos:pos] + n.Indices[pos+1:] // rest without char at 'pos'
	}

	return newPos
}

// addRoute adds a node with the given handle to the path.
// Not concurrency-safe!
func (n *node) addRoute(path string, cInfo *controllerInfo) {
	fullPath := path
	n.Priority++
	numParams := countParams(path)

	// non-empty tree
	if len(n.Path) > 0 || len(n.Children) > 0 {
	walk:
		for {
			// Update maxParams of the current node
			if numParams > n.MaxParams {
				n.MaxParams = numParams
			}

			// Find the longest common prefix.
			// This also implies that the common prefix contains no ':' or '*'
			// since the existing key can't contain those chars.
			i := 0
			max := min(len(path), len(n.Path))
			for i < max && path[i] == n.Path[i] {
				i++
			}

			// Split edge
			if i < len(n.Path) {
				child := node{
					Path:      n.Path[i:],
					WildChild: n.WildChild,
					Indices:   n.Indices,
					Children:  n.Children,
					CtrlInfo:     n.CtrlInfo,
					Priority:  n.Priority - 1,
				}

				// Update maxParams (max of all children)
				for i := range child.Children {
					if child.Children[i].MaxParams > child.MaxParams {
						child.MaxParams = child.Children[i].MaxParams
					}
				}

				n.Children = []*node{&child}
				// []byte for proper unicode char conversion, see #65
				n.Indices = string([]byte{n.Path[i]})
				n.Path = path[:i]
				n.CtrlInfo = nil
				n.WildChild = false
			}

			// Make new node a child of this node
			if i < len(path) {
				path = path[i:]

				if n.WildChild {
					n = n.Children[0]
					n.Priority++

					// Update maxParams of the child node
					if numParams > n.MaxParams {
						n.MaxParams = numParams
					}
					numParams--

					// Check if the wildcard matches
					if len(path) >= len(n.Path) && n.Path == path[:len(n.Path)] {
						// check for longer wildcard, e.g. :name and :names
						if len(n.Path) >= len(path) || path[len(n.Path)] == '/' {
							continue walk
						}
					}

					panic("path segment '" + path +
						"' conflicts with existing wildcard '" + n.Path +
						"' in path '" + fullPath + "'")
				}

				c := path[0]

				// slash after param
				if n.NodeType == param && c == '/' && len(n.Children) == 1 {
					n = n.Children[0]
					n.Priority++
					continue walk
				}

				// Check if a child with the next path byte exists
				for i = 0; i < len(n.Indices); i++ {
					if c == n.Indices[i] {
						i = n.incrementChildPrio(i)
						n = n.Children[i]
						continue walk
					}
				}

				// Otherwise insert it
				if c != ':' && c != '*' {
					// []byte for proper unicode char conversion, see #65
					n.Indices += string([]byte{c})
					child := &node{
						MaxParams: numParams,
					}
					n.Children = append(n.Children, child)
					n.incrementChildPrio(len(n.Indices) - 1)
					n = child
				}
				n.insertChild(numParams, path, fullPath, cInfo)
				return

			} else if i == len(path) { // Make node a (in-path) leaf
				if n.CtrlInfo != nil {
					panic("a handle is already registered for path '" + fullPath + "'")
				}
				n.CtrlInfo = cInfo
			}
			return
		}
	} else { // Empty tree
		n.insertChild(numParams, path, fullPath, cInfo)
		n.NodeType = root
	}
}

func (n *node) insertChild(numParams uint8, path, fullPath string, handle *controllerInfo) {
	var offset int // already handled bytes of the path

	// find prefix until first wildcard (beginning with ':'' or '*'')
	for i, max := 0, len(path); numParams > 0; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}

		// find wildcard end (either '/' or path end)
		end := i + 1
		for end < max && path[end] != '/' {
			switch path[end] {
			// the wildcard name must not contain ':' and '*'
			case ':', '*':
				panic("only one wildcard per path segment is allowed, has: '" +
					path[i:] + "' in path '" + fullPath + "'")
			default:
				end++
			}
		}

		// check if this Node existing children which would be
		// unreachable if we insert the wildcard here
		if len(n.Children) > 0 {
			panic("wildcard route '" + path[i:end] +
				"' conflicts with existing children in path '" + fullPath + "'")
		}

		// check if the wildcard has a name
		if end-i < 2 {
			panic("wildcards must be named with a non-empty name in path '" + fullPath + "'")
		}

		if c == ':' { // param
			// split path at the beginning of the wildcard
			if i > 0 {
				n.Path = path[offset:i]
				offset = i
			}

			child := &node{
				NodeType:     param,
				MaxParams: numParams,
			}
			n.Children = []*node{child}
			n.WildChild = true
			n = child
			n.Priority++
			numParams--

			// if the path doesn't end with the wildcard, then there
			// will be another non-wildcard subpath starting with '/'
			if end < max {
				n.Path = path[offset:end]
				offset = end

				child := &node{
					MaxParams: numParams,
					Priority:  1,
				}
				n.Children = []*node{child}
				n = child
			}

		} else { // catchAll
			if end != max || numParams > 1 {
				panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
			}

			if len(n.Path) > 0 && n.Path[len(n.Path)-1] == '/' {
				panic("catch-all conflicts with existing handle for the path segment root in path '" + fullPath + "'")
			}

			// currently fixed width 1 for '/'
			i--
			if path[i] != '/' {
				panic("no / before catch-all in path '" + fullPath + "'")
			}
			paramName := path[strings.Index(path, "*")+1:]
			if paramName != "pathInfo" {
				panic("the parameter name of the catch-all rule should be 'pathInfo': '" + fullPath + "'")
			}

			n.Path = path[offset:i]
			// first node: catchAll node with empty path
			child := &node{
				WildChild: true,
				NodeType:     catchAll,
				MaxParams: 1,
			}
			n.Children = []*node{child}
			n.Indices = string(path[i])
			n = child
			n.Priority++

			// second node: node holding the variable
			child = &node{
				Path:      path[i:],
				NodeType:     catchAll,
				MaxParams: 1,
				CtrlInfo:     handle,
				Priority:  1,
			}

			n.Children = []*node{child}

			return
		}
	}

	// insert remaining path part and handle to the leaf
	n.Path = path[offset:]
	n.CtrlInfo = handle
}

// Returns the handle registered with the given path (key). The values of
// wildcards are saved to a map.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
func (n *node) getValue(path string) (cInfo *controllerInfo, p RouteData, tsr bool) {
walk: // Outer loop for walking the tree
	for {
		if len(path) > len(n.Path) {
			if path[:len(n.Path)] == n.Path {
				path = path[len(n.Path):]
				// If this node does not have a wildcard (param or catchAll)
				// child,  we can just look up the next child node and continue
				// to walk down the tree
				if !n.WildChild {
					c := path[0]
					for i := 0; i < len(n.Indices); i++ {
						if c == n.Indices[i] {
							n = n.Children[i]
							continue walk
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					tsr = (path == "/" && n.CtrlInfo != nil)
					return

				}

				// handle wildcard child
				n = n.Children[0]
				switch n.NodeType {
				case param:
					// find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// save param value
					if p == nil {
						// lazy allocation
						p = make(RouteData, 0, n.MaxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].Key = n.Path[1:]
					p[i].Value = path[:end]

					// we need to go deeper!
					if end < len(path) {
						if len(n.Children) > 0 {
							path = path[end:]
							n = n.Children[0]
							continue walk
						}

						// ... but we can't
						tsr = (len(path) == end+1)
						return
					}

					if cInfo = n.CtrlInfo; cInfo != nil {
						return
					} else if len(n.Children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						n = n.Children[0]
						tsr = (n.Path == "/" && n.CtrlInfo != nil)
					}

					return

				case catchAll:
					// save param value
					if p == nil {
						// lazy allocation
						p = make(RouteData, 0, n.MaxParams)
					}
					i := len(p)
					p = p[:i+1] // expand slice within preallocated capacity
					p[i].Key = n.Path[2:]
					p[i].Value = path

					cInfo = n.CtrlInfo
					return

				default:
					panic("invalid node type")
				}
			}
		} else if path == n.Path {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if cInfo = n.CtrlInfo; cInfo != nil {
				return
			}

			if path == "/" && n.WildChild && n.NodeType != root {
				tsr = true
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i := 0; i < len(n.Indices); i++ {
				if n.Indices[i] == '/' {
					n = n.Children[i]
					tsr = (len(n.Path) == 1 && n.CtrlInfo != nil) ||
						(n.NodeType == catchAll && n.Children[0].CtrlInfo != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		tsr = (path == "/") ||
			(len(n.Path) == len(path)+1 && n.Path[len(path)] == '/' &&
				path == n.Path[:len(n.Path)-1] && n.CtrlInfo != nil)
		return
	}
}

// Makes a case-insensitive lookup of the given path and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) (ciPath []byte, found bool) {
	ciPath = make([]byte, 0, len(path)+1) // preallocate enough memory

	// Outer loop for walking the tree
	for len(path) >= len(n.Path) && strings.ToLower(path[:len(n.Path)]) == strings.ToLower(n.Path) {
		path = path[len(n.Path):]
		ciPath = append(ciPath, n.Path...)

		if len(path) > 0 {
			// If this node does not have a wildcard (param or catchAll) child,
			// we can just look up the next child node and continue to walk down
			// the tree
			if !n.WildChild {
				r := unicode.ToLower(rune(path[0]))
				for i, index := range n.Indices {
					// must use recursive approach since both index and
					// ToLower(index) could exist. We must check both.
					if r == unicode.ToLower(index) {
						out, found := n.Children[i].findCaseInsensitivePath(path, fixTrailingSlash)
						if found {
							return append(ciPath, out...), true
						}
					}
				}

				// Nothing found. We can recommend to redirect to the same URL
				// without a trailing slash if a leaf exists for that path
				found = (fixTrailingSlash && path == "/" && n.CtrlInfo != nil)
				return
			}

			n = n.Children[0]
			switch n.NodeType {
			case param:
				// find param end (either '/' or path end)
				k := 0
				for k < len(path) && path[k] != '/' {
					k++
				}

				// add param value to case insensitive path
				ciPath = append(ciPath, path[:k]...)

				// we need to go deeper!
				if k < len(path) {
					if len(n.Children) > 0 {
						path = path[k:]
						n = n.Children[0]
						continue
					}

					// ... but we can't
					if fixTrailingSlash && len(path) == k+1 {
						return ciPath, true
					}
					return
				}

				if n.CtrlInfo != nil {
					return ciPath, true
				} else if fixTrailingSlash && len(n.Children) == 1 {
					// No handle found. Check if a handle for this path + a
					// trailing slash exists
					n = n.Children[0]
					if n.Path == "/" && n.CtrlInfo != nil {
						return append(ciPath, '/'), true
					}
				}
				return

			case catchAll:
				return append(ciPath, path...), true

			default:
				panic("invalid node type")
			}
		} else {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.CtrlInfo != nil {
				return ciPath, true
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash {
				for i := 0; i < len(n.Indices); i++ {
					if n.Indices[i] == '/' {
						n = n.Children[i]
						if (len(n.Path) == 1 && n.CtrlInfo != nil) ||
							(n.NodeType == catchAll && n.Children[0].CtrlInfo != nil) {
							return append(ciPath, '/'), true
						}
						return
					}
				}
			}
			return
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash {
		if path == "/" {
			return ciPath, true
		}
		if len(path)+1 == len(n.Path) && n.Path[len(path)] == '/' &&
			strings.ToLower(path) == strings.ToLower(n.Path[:len(path)]) &&
			n.CtrlInfo != nil {
			return append(ciPath, n.Path...), true
		}
	}
	return
}
