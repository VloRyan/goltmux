package goltmux

import (
	"net/http"
	"strings"
)

type Router struct {
	tree            RouteTree
	NotFoundHandler http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	contentType, _ := extractContentType(req)
	handler, params := r.Lookup(contentType, req.Method, req.URL.Path)
	if handler != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
		handler(w, req)
	} else {
		if r.NotFoundHandler != nil {
			r.NotFoundHandler(w, req)
		} else {
			http.NotFound(w, req)
		}
	}
}

func extractContentType(req *http.Request) (string, string) {
	header := req.Header.Get("Content-Type")
	paramStart := strings.Index(header, ";")
	if paramStart == -1 {
		return header, ""
	}
	return header[:paramStart], strings.TrimSpace(header[paramStart+1:])
}

func (r *Router) Lookup(contentType, method, url string) (http.HandlerFunc, map[string]string) {
	path := r.makePathSlice(contentType, method, url)
	elem, param := r.tree.Resolve(path)
	if elem == nil {
		return r.NotFoundHandler, nil
	}
	return elem.HandleRouteFunc, param
}

func (r *Router) Handle(contentType, method, url string, handler http.HandlerFunc) {
	path := r.makePathSlice(contentType, method, url)
	node, err := r.tree.Add(path)
	if err != nil {
		panic(err)
	}
	node.HandleRouteFunc = handler
}

func (r *Router) makePathSlice(contentType, method, url string) []string {
	parts := strings.Split(url, "/")
	cleanParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		cleanParts = append(cleanParts, part)
	}
	path := make([]string, len(cleanParts)+2)
	path[0] = contentType
	path[1] = method
	for i, part := range cleanParts {
		path[i+2] = part
	}
	return path
}
