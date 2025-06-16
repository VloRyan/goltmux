package goltmux

import (
	"net/http"
	"reflect"
	"testing"
)

var dummyHandlerFunc = http.NotFound

func TestRoutePathElement_Resolve(t *testing.T) {
	tests := []struct {
		name       string
		elem       RoutePathElement
		path       string
		wantMatch  bool
		wantParams map[string]string
	}{{
		name: "GIVEN matching path THEN return element",
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
		name: "GIVEN parameter marker THEN return element and params",
		elem: RoutePathElement{
			Path: ":param",
			Children: []*RoutePathElement{{
				Path: "*",
				Children: []*RoutePathElement{{
					Path:            ":param2",
					HandleRouteFunc: dummyHandlerFunc,
				}},
			}},
		},
		path:       "/myParam/path/anotherParam",
		wantMatch:  true,
		wantParams: map[string]string{":param": "myParam", ":param2": "anotherParam"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, params := tt.elem.Resolve(tt.path); (got != nil) != tt.wantMatch {
				t.Errorf("Resolve() = %v, want %v", got != nil, tt.wantMatch)
			} else {
				if tt.wantParams == nil {
					if len(params) != 0 {
						t.Errorf("Resolve() params = %v, want %v", params, make(map[string]string))
					}
				} else {
					if !reflect.DeepEqual(tt.wantParams, params) {
						t.Errorf("Resolve() params = %v, want:%v", params, tt.wantParams)
					}
				}
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
					}},
				}},
			}},
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
					},
				}},
			}},
		},
		paths: []string{"/domain/item/:id", "/domain/item/other"},
	}, {
		name: "GIVEN path trailing slash THEN add children",
		want: RoutePathElement{
			Children: []*RoutePathElement{{
				Path: "domain", Children: []*RoutePathElement{{
					Path: "item",
				}},
			}},
		},
		paths: []string{"/domain/item/"},
	}, {
		name: "GIVEN path with double slash THEN add children",
		want: RoutePathElement{
			Children: []*RoutePathElement{{
				Path: "domain", Children: []*RoutePathElement{{
					Path: "item",
				}},
			}},
		},
		paths: []string{"/domain//item/", "//domain/item/", "/domain/item//", "/domain/item///"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := RoutePathElement{}
			for _, path := range tt.paths {
				_, _ = e.Add(path)
			}

			if !reflect.DeepEqual(tt.want, e) {
				t.Errorf("Add() mismatch:\nwant:%v, got:%v", tt.want, e)
			}
		})
	}
}
