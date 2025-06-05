package goltmux

import (
	"errors"
	"net/http"
	"slices"
	"strings"
)

type HandleRouteFunc func(method, path string, handler http.HandlerFunc)
type RouteElement interface {
	Resolve(path string) (RouteElement, map[string]string)
	IsWildcard() bool
	IsParam() bool
	Add(path string) (RouteElement, error)
	Handler() http.HandlerFunc
	UpdateHandler(fun http.HandlerFunc)
	Walk(func(child RouteElement) bool)
}
type RoutePathElement struct {
	Path            string
	HandleRouteFunc http.HandlerFunc
	Children        []*RoutePathElement
}

func (r *RoutePathElement) Resolve(path string) (RouteElement, map[string]string) {
	params := make(map[string]string)
	return r.resolve(path, params), params
}
func (r *RoutePathElement) resolve(path string, params map[string]string) RouteElement {
	path = strings.TrimPrefix(path, "/")
	pathElem := path
	firstSlash := strings.Index(path, "/")
	if firstSlash != -1 {
		pathElem = path[:firstSlash]
	}
	if !r.IsWildcard() && r.Path != pathElem {
		return nil
	}
	if r.IsParam() {
		params[r.Path] = pathElem
	}
	childPath := path[len(pathElem):]
	if childPath == "" {
		return r
	}
	for _, child := range r.Children {
		if found := child.resolve(childPath, params); found != nil {
			return found
		}
	}
	return nil
}

func (r *RoutePathElement) IsWildcard() bool {
	return r.Path == "*" || r.IsParam()
}
func (r *RoutePathElement) IsParam() bool {
	return strings.HasPrefix(r.Path, ":")
}

func (r *RoutePathElement) Add(path string) (RouteElement, error) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	currentElem := r
	for i, part := range parts {
		if part == "" {
			continue
		}
		found := false
		isPlaceholder := strings.HasPrefix(part, ":")
		for _, elem := range currentElem.Children {
			if elem.Path == part {
				currentElem = elem
				found = true
				break
			}
		}
		if !found {
			for j := i; j < len(parts); j++ {
				if parts[j] == "" {
					continue
				}
				newElem := &RoutePathElement{
					Path: parts[j],
				}
				if i == j &&
					len(currentElem.Children) > 0 &&
					currentElem.Children[len(currentElem.Children)-1].IsWildcard() {
					if isPlaceholder {
						return nil, errors.New("path contains placeholder")
					}
					currentElem.Children = slices.Insert(currentElem.Children, len(currentElem.Children)-1, newElem)
				} else {
					currentElem.Children = append(currentElem.Children, newElem)
				}
				currentElem = newElem
			}
			break
		}
	}
	return currentElem, nil
}

func (r *RoutePathElement) Handler() http.HandlerFunc {
	return r.HandleRouteFunc
}
func (r *RoutePathElement) UpdateHandler(fun http.HandlerFunc) {
	r.HandleRouteFunc = fun
}

func (r *RoutePathElement) Walk(f func(child RouteElement) bool) {
	for _, child := range r.Children {
		if !f(child) {
			break
		}
	}
}
