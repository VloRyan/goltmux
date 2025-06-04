package goltmux

import (
	"errors"
	"net/http"
	"reflect"
	"slices"
	"strings"
)

type HandleRouteFunc func(method, path string, handler http.HandlerFunc)
type RouteElement interface {
	Resolve(path string) RouteElement
	IsWildcard() bool
	Add(path string) (RouteElement, error)
	Handler() http.HandlerFunc
	UpdateHandler(fun http.HandlerFunc)
	Walk(func(child RouteElement))
}
type RoutePathElement struct {
	Path            string
	HandleRouteFunc http.HandlerFunc
	Children        []*RoutePathElement
}

func (r *RoutePathElement) Resolve(path string) RouteElement {
	path = strings.TrimPrefix(path, "/")
	pathElem := path
	firstSlash := strings.Index(path, "/")
	if firstSlash != -1 {
		pathElem = path[:firstSlash]
	}
	if !r.IsWildcard() && r.Path != pathElem {
		return nil
	}
	childPath := ""
	if firstSlash != -1 {
		childPath = path[firstSlash+1:]
	} else {
		childPath = path[len(pathElem):]
	}
	if childPath == "" {
		return r
	}
	for _, child := range r.Children {
		if found := child.Resolve(childPath); found != nil {
			return found
		}
	}
	return nil
}

func (r *RoutePathElement) IsWildcard() bool {
	return r.Path == "*" || strings.HasPrefix(r.Path, ":")
}

func (r *RoutePathElement) Add(path string) (RouteElement, error) {
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	currentElem := r
	for i, part := range parts {
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

func (r *RoutePathElement) Walk(f func(child RouteElement)) {
	for _, child := range r.Children {
		f(child)
	}
}

type RouteRootElement struct {
	Children []*RoutePathElement
}

func (r *RouteRootElement) Walk(f func(child RouteElement)) {
	for _, child := range r.Children {
		f(child)
	}
}

func (r *RouteRootElement) Resolve(path string) RouteElement {
	for _, child := range r.Children {
		if elem := child.Resolve(path); !reflect.ValueOf(elem).IsNil() {
			return elem
		}
	}
	return nil
}

func (r *RouteRootElement) IsWildcard() bool {
	return false
}

func (r *RouteRootElement) Add(path string) (RouteElement, error) {
	child := r.Resolve(path)
	if child == nil || reflect.ValueOf(child).IsNil() {
		newChild := &RoutePathElement{
			Path: path,
		}
		r.Children = append(r.Children, newChild)
		return newChild, nil
	}
	return child.Add(path)
}

func (r *RouteRootElement) Handler() http.HandlerFunc {
	return nil
}

func (r *RouteRootElement) UpdateHandler(_ http.HandlerFunc) {

}
