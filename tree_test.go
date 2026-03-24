package goltmux

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

var dummyHandlerFunc = http.NotFound

func TestRouteTree_Resolve(t *testing.T) {
	tests := []struct {
		name       string
		node       RouteNode
		path       string
		wantMatch  bool
		wantParams map[string]string
	}{{
		name: "GIVEN matching path THEN return element",
		node: RouteNode{
			Path:            "test",
			HandleRouteFunc: dummyHandlerFunc,
		},
		path:      "test",
		wantMatch: true,
	}, {
		name: "GIVEN non matching path THEN return nil",
		node: RouteNode{
			Path:            "test",
			HandleRouteFunc: dummyHandlerFunc,
		},
		path:      "test1",
		wantMatch: false,
	}, {
		name: "GIVEN parameter marker THEN return element and params",
		node: RouteNode{
			Path: ":param",
			Children: []*RouteNode{{
				Path: "path",
				Children: []*RouteNode{{
					Path:            ":param2",
					HandleRouteFunc: dummyHandlerFunc,
				}},
			}},
		},
		path:       "myParam/path/anotherParam",
		wantMatch:  true,
		wantParams: map[string]string{":param": "myParam", ":param2": "anotherParam"},
	}, {
		name: "GIVEN multi parameter marker THEN return element and matching params",
		node: RouteNode{
			Path: "path",
			Children: []*RouteNode{{
				Path: ":id",
				Children: []*RouteNode{{
					Path:            "detail",
					HandleRouteFunc: dummyHandlerFunc,
				}},
			}, {
				Path: ":other_id",
			}},
		},
		path:       "path/7/detail",
		wantMatch:  true,
		wantParams: map[string]string{":id": "7"},
	}, {
		name: "GIVEN multi parameter marker THEN return element and matching params",
		node: RouteNode{
			Path: "path",
			Children: []*RouteNode{{
				Path: ":id",
				Children: []*RouteNode{{
					Path:            "detail",
					HandleRouteFunc: dummyHandlerFunc,
				}},
			}, {
				Path: ":other_id",
			}},
		},
		path:       "path/7",
		wantMatch:  true,
		wantParams: map[string]string{":other_id": "7"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &RouteTree{Root: RouteNode{Children: []*RouteNode{&tt.node}}}
			if got, params := tree.Resolve(strings.Split(tt.path, "/")); (got != nil) != tt.wantMatch {
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

func TestRouteTree_Add(t *testing.T) {
	tests := []struct {
		name    string
		paths   []string
		want    RouteNode
		wantErr []error
	}{{
		name: "GIVEN plain path THEN add children",
		want: RouteNode{
			Children: []*RouteNode{{
				Path: "test", Children: []*RouteNode{{
					Path: "this", Children: []*RouteNode{{
						Path: "func",
					}},
				}},
			}},
		},
		paths: []string{"test/this/func"},
	}, {
		name: "GIVEN path with param THEN param child at end",
		want: RouteNode{
			Children: []*RouteNode{{
				Path: "domain", Children: []*RouteNode{{
					Path: "item", Children: []*RouteNode{
						{Path: "other"},
						{Path: ":id"},
					},
				}},
			}},
		},
		paths: []string{"domain/item/:id", "domain/item/other"},
	}, {
		name: "GIVEN paths with placeholders THEN add children",
		want: RouteNode{
			Children: []*RouteNode{{
				Path: "domain", Children: []*RouteNode{
					{
						Path: "object", Children: []*RouteNode{
							{Path: ":id"},
							{
								Path: ":objectId", Children: []*RouteNode{
									{
										Path: "status", Children: []*RouteNode{
											{Path: ":id"},
										},
									},
								},
							},
						},
					},
				},
			}},
		},
		paths: []string{"domain/object/:id", "domain/object/:objectId/status/:id"},
	}, {
		name:    "GIVEN paths with ambiguous params THEN error",
		want:    RouteNode{Children: []*RouteNode{{Path: "domain"}}},
		paths:   []string{"domain/:id/:test"},
		wantErr: []error{ErrPathContainsAmbiguousParams},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := RouteTree{}
			for i, path := range tt.paths {
				if _, err := tree.Add(strings.Split(path, "/")); len(tt.wantErr) > i && !errors.Is(err, tt.wantErr[i]) {
					t.Fatalf("Add() mismatch:\nwant error:%v, got:%v", tt.wantErr, err)
				}
			}

			if !reflect.DeepEqual(tt.want, tree.Root) {
				t.Errorf("Add() mismatch:\nwant:%v, got:%v", tt.want, tree.Root)
			}
		})
	}
}
