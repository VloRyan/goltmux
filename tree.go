package goltmux

import (
	"errors"
	"net/http"
	"strings"
)

func IsPlaceholder(element string) bool {
	return strings.HasPrefix(element, ":")
}

type RouteNode struct {
	Parent          *RouteNode
	Path            string
	HandleRouteFunc http.HandlerFunc
	Children        []*RouteNode
}

var ErrPathCanOnlyContainOnePlaceholderPerLevel = errors.New("path can only contain one placeholder per level")

func (r *RouteNode) Matches(path string) (bool, string) {
	if r.IsPlaceholder() {
		return true, r.Path
	}
	return r.Path == path, r.Path
}

func (r *RouteNode) IsPlaceholder() bool {
	return IsPlaceholder(r.Path)
}

type RouteTree struct {
	Root RouteNode
}

func (t *RouteTree) Resolve(path []string) (*RouteNode, map[string]string) {
	params := make(map[string]string)
	currentNode := &t.Root
	var nextNode *RouteNode
	for _, p := range path {
		if currentNode == nil {
			break
		}
		nextNode = nil
		for _, child := range currentNode.Children {
			match, matchedPath := child.Matches(p)
			if !match {
				continue
			}
			nextNode = child
			if child.IsPlaceholder() {
				if len(matchedPath) > 1 {
					params[matchedPath] = p
				}
			}
			break
		}
		currentNode = nextNode
	}
	if currentNode != nil {
		return currentNode, params
	}
	return nil, nil
}

func (t *RouteTree) Add(path []string) (*RouteNode, error) {
	currentNode := &t.Root
	for _, p := range path {
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
			newNode := &RouteNode{
				Parent: currentNode,
				Path:   p,
			}
			lastNodeIsPlaceholder := len(currentNode.Children) > 0 && currentNode.Children[len(currentNode.Children)-1].IsPlaceholder()
			if newNode.IsPlaceholder() {
				if lastNodeIsPlaceholder {
					return nil, ErrPathCanOnlyContainOnePlaceholderPerLevel
				}
			}
			if lastNodeIsPlaceholder {
				placeholderNode := currentNode.Children[len(currentNode.Children)-1]
				currentNode.Children[len(currentNode.Children)-1] = newNode
				currentNode.Children = append(currentNode.Children, placeholderNode)
			} else {
				currentNode.Children = append(currentNode.Children, newNode)
			}
			currentNode = newNode
		}
	}
	return currentNode, nil
}
