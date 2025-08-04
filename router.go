package goltmux

import (
	"net/http"
	"strings"
)

type Router struct {
	root            RouteElement
	NotFoundHandler http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		root: &RoutePathElement{},
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, params := r.Lookup(req.Header.Get("Content-Type"), req.Method, req.URL.Path)
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

func (r *Router) Lookup(contentType, method, url string) (http.HandlerFunc, map[string]string) {
	contentType = strings.ReplaceAll(contentType, "/", "_")
	var handler http.HandlerFunc
	params := make(map[string]string)
	r.root.Walk(func(child RouteElement) bool {
		if ctRoute, p := child.Resolve(contentType); ctRoute != nil {
			for k, v := range p {
				params[k] = v
			}
			ctRoute.Walk(func(child RouteElement) bool {
				if mRoute, p := child.Resolve(method); mRoute != nil {
					for k, v := range p {
						params[k] = v
					}
					mRoute.Walk(func(child RouteElement) bool {
						if uRoute, p := child.Resolve(url); uRoute != nil {
							for k, v := range p {
								params[k] = v
							}
							handler = uRoute.Handler()
							if handler != nil {
								return false
							}
						}
						return true
					})
					if handler != nil {
						return false
					}
					if mRoute.Handler() != nil {
						handler = mRoute.Handler()
						return false
					}
					return true
				}
				if handler != nil {
					return false
				}
				return true
			})
			if handler != nil {
				return false
			}
			if ctRoute.Handler() != nil {
				handler = ctRoute.Handler()
				return false
			}
			return true
		}
		if handler != nil {
			return false
		}
		return true
	})
	return handler, params
}

func (r *Router) HandleMethod(method string, path string, handler http.HandlerFunc) {
	r.Handle("*", method, path, handler)
}

func (r *Router) Handle(contentType, method, path string, handler http.HandlerFunc) {
	ctRoute, err := r.root.Add(strings.ReplaceAll(contentType, "/", "_"))
	if err != nil {
		panic(err)
	}
	mRoute, err := ctRoute.Add(method)
	if err != nil {
		panic(err)
	}
	uRoute, err := mRoute.Add(path)
	if err != nil {
		panic(err)
	}
	if uRoute.Handler() != nil {
		panic("handler for path " + path + " already defined")
	}
	uRoute.UpdateHandler(handler)
}

func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.HandleMethod(http.MethodGet, path, handler)
}
