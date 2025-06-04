package goltmux

import (
	"net/http"
	"reflect"
	"testing"
)

var dummyHandlerFunc = http.NotFound

func TestRoutePathElement_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		elem      RoutePathElement
		path      string
		wantMatch bool
	}{{
		name: "GIVEN matching path THEN return handler",
		elem: RoutePathElement{
			Path:            "test",
			HandleRouteFunc: dummyHandlerFunc,
		},
		path:      "/test",
		wantMatch: true,
	}, {
		name: "GIVEN non matching path THEN return nil",
		elem: RoutePathElement{
			Path:            "test",
			HandleRouteFunc: dummyHandlerFunc,
		},
		path:      "/test1",
		wantMatch: false,
	}, {
		name: "GIVEN parameter marker THEN handler",
		elem: RoutePathElement{
			Path:            ":param",
			HandleRouteFunc: dummyHandlerFunc,
		},
		path:      "/myParam",
		wantMatch: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.elem.Resolve(tt.path); (got != nil) != tt.wantMatch {
				t.Errorf("Resolve() = %v, want %v", got != nil, tt.wantMatch)
			}
		})
	}
}
func TestRoutePathElement_Add(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		want  RoutePathElement
	}{{
		name: "GIVEN plain path THEN add children",
		want: RoutePathElement{
			Children: []*RoutePathElement{{
				Path: "test", Children: []*RoutePathElement{{
					Path: "this", Children: []*RoutePathElement{{
						Path: "func",
					}}}}}},
		},
		paths: []string{"/test/this/func"},
	}, {
		name: "GIVEN path with param THEN param child at end",
		want: RoutePathElement{
			Children: []*RoutePathElement{{
				Path: "domain", Children: []*RoutePathElement{{
					Path: "item", Children: []*RoutePathElement{
						{Path: "other"},
						{Path: ":id"},
					}}}}},
		},
		paths: []string{"/domain/item/:id", "/domain/item/other"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := RoutePathElement{}
			for _, path := range tt.paths {
				_, _ = e.Add(path)
			}

			if !reflect.DeepEqual(tt.want, e) {
				t.Errorf("ServeHTTP() mismatch:\nwant:%v, got:%v", tt.want, e)
			}
		})
	}
}
