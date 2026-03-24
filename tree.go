package goltmux

import (
	"errors"
	"net/http"
	"sort"
	"strings"
)

type RouteNode struct {
	Path            string
	HandleRouteFunc http.HandlerFunc
	Children        []*RouteNode
}

var (
	ErrPathWildcardMustBeLeaf      = errors.New("wildcards can only be leafs")
	ErrPathContainsAmbiguousParams = errors.New("path contains ambiguous parameter marker")
)

func (r *RouteNode) Matches(path string) (bool, string) {
	if r.IsPlaceholder() {
		return true, r.Path
	}
	return r.Path == path, r.Path
}

func (r *RouteNode) IsLeaf() bool {
	return len(r.Children) == 0
}

func (r *RouteNode) IsPlaceholder() bool {
	return strings.HasPrefix(r.Path, ":")
}

func (r *RouteNode) IsNamedPlaceholder() bool {
	return r.IsPlaceholder() && len(r.Path) > 1
}

type RouteTree struct {
	Root RouteNode
}

func (t *RouteTree) Resolve(path []string) (*RouteNode, map[string]string) {
	params := make(map[string]string)
	currentNode := &t.Root
	var nextNode *RouteNode = nil
	lastIndex := len(path) - 1
	for i, p := range path {
		if currentNode == nil {
			break
		}
		lastToken := i == lastIndex
		nextNode = nil
		for _, child := range currentNode.Children {
			match, matchedPath := child.Matches(p)
			if !match || lastToken != child.IsLeaf() {
				continue
			}
			nextNode = child
			if child.IsPlaceholder() {
				if !lastToken {
					nextToken := path[i+1]
					var childMatch bool
					for _, childOfChild := range child.Children {
						if childMatch, _ = childOfChild.Matches(nextToken); !childMatch {
							continue
						}
					}
					if !childMatch {
						continue
					}
				}
				if len(matchedPath) > 1 {
					params[matchedPath] = p
				}
			}
			break
		}
		currentNode = nextNode
	}
	if currentNode != nil && currentNode.IsLeaf() {
		return currentNode, params
	}
	return nil, nil
}

func (t *RouteTree) Add(path []string) (*RouteNode, error) {
	currentNode := &t.Root
	for i, p := range path {
		if p == "" {
			continue
		}
		found := false
		for _, node := range currentNode.Children {
			if node.Path == p {
				currentNode = node
				found = true
				break
			}
		}
		if !found {
			newElem := &RouteNode{
				Path: p,
			}
			if newElem.IsPlaceholder() && i < len(path)-1 {
				nextElem := &RouteNode{
					Path: path[i+1],
				}
				if nextElem.IsPlaceholder() {
					return nil, ErrPathContainsAmbiguousParams
				}
			}
			currentNode.Children = append(currentNode.Children, newElem)
			sort.SliceStable(currentNode.Children, func(e, e2 int) bool {
				if currentNode.Children[e].IsPlaceholder() != currentNode.Children[e2].IsPlaceholder() {
					return currentNode.Children[e2].IsPlaceholder()
				}
				return currentNode.Children[e].Path < currentNode.Children[e2].Path
			})
			currentNode = newElem
		}
	}
	return currentNode, nil
}
