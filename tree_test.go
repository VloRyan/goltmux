package goltmux

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRouteTree_AddAndResolve(t *testing.T) {
	tests := []struct {
		name            string
		registerPath    [][]string
		path            string
		wantMatchedPath string
		wantParams      map[string]string
	}{{
		name:            "GIVEN matching path THEN return element",
		registerPath:    [][]string{{"test"}},
		path:            "test",
		wantMatchedPath: "test",
	}, {
		name:         "GIVEN non matching path THEN return nil",
		registerPath: [][]string{{"test"}},
		path:         "test1",
	}, {
		name:            "GIVEN parameter marker THEN return element with params",
		registerPath:    [][]string{{":param", "path", ":param2"}},
		path:            "myParam/path/anotherParam",
		wantMatchedPath: ":param/path/:param2",
		wantParams:      map[string]string{":param": "myParam", ":param2": "anotherParam"},
	}, {
		name: "GIVEN multi parameter marker THEN return element with matching params",
		registerPath: [][]string{
			{"path", ":id", "detail"},
			{"path", ":id"},
		},
		path:            "path/7",
		wantMatchedPath: "path/:id",
		wantParams:      map[string]string{":id": "7"},
	}, {
		name: "GIVEN multi parameter marker THEN return sub element with matching params",
		registerPath: [][]string{
			{"path", ":id"},
			{"path", ":id", "detail"},
		},
		path:            "path/7/detail",
		wantMatchedPath: "path/:id/detail",
		wantParams:      map[string]string{":id": "7"},
	}, {
		name: "GIVEN multi parameter marker THEN return element and matching params2",
		registerPath: [][]string{
			{"application_json", "POST", "test"},
			{"application_json", "GET", "test"},
			{"application_json", "DELETE", "test"},
		},
		path:            "application_json/GET/test",
		wantMatchedPath: "application_json/GET/test",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := &RouteTree{Root: RouteNode{}}
			for _, p := range tt.registerPath {
				_, err := tree.Add(p)
				if err != nil {
					t.Fatal(err)
				}
			}
			got, params := tree.Resolve(strings.Split(tt.path, "/"))

			if got == nil {
				if len(tt.wantMatchedPath) != 0 {
					t.Errorf("Resolve() = %v, want %v", got != nil, tt.wantMatchedPath)
				}
				return
			}
			matchedPath := travelUpPath(got)
			if matchedPath != tt.wantMatchedPath {
				t.Errorf("Resolve() matchedPath= %v, wantMatchedPath=%v", matchedPath, tt.wantMatchedPath)
			}
			if tt.wantParams == nil {
				if len(params) != 0 {
					t.Errorf("Resolve() params = %v, want %v", params, make(map[string]string))
				}
			} else {
				if !reflect.DeepEqual(tt.wantParams, params) {
					t.Errorf("Resolve() params = %v, want:%v", params, tt.wantParams)
				}
			}
		})
	}
}

func TestRouteTree_AddConstraints(t *testing.T) {
	tests := []struct {
		name         string
		registerPath [][]string
		want         error
	}{{
		name: "GIVEN paths with multi params in one level THEN ErrPathCanOnlyContainOnePlaceholderPerLevel",
		registerPath: [][]string{
			{"domain", "correct", "test"},
			{"domain", ":id", "test"},
			{"domain", "other_correct", "test"},
			{"domain", ":another_id", "test"},
		},
		want: ErrPathCanOnlyContainOnePlaceholderPerLevel,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := RouteTree{}
			gotErr := false
			for _, p := range tt.registerPath {
				if _, err := tree.Add(p); err != nil {
					if errors.Is(err, tt.want) {
						gotErr = true
					} else {
						t.Fatalf("Add():\nwant error:%v, got:%v", tt.want, err)
					}
				}
			}
			if !gotErr {
				t.Errorf("Add():\nwant:%v, got: NIL", tt.want)
			}
		})
	}
}

func travelUpPath(node *RouteNode) string {
	path := node.Path
	currentNode := node.Parent
	for currentNode != nil && currentNode.Parent != nil {
		path = currentNode.Path + "/" + path
		currentNode = currentNode.Parent
	}
	return path
}
