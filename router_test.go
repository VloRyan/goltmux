package goltmux

import (
	"net/http"
	"reflect"
	"testing"
)

type mockWriter struct {
	statusCode int
	body       []byte
}

func (m *mockWriter) Header() http.Header {
	return http.Header{}
}

func (m *mockWriter) Write(bytes []byte) (int, error) {
	m.body = bytes
	return len(bytes), nil
}

func (m *mockWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}
func writeSuccess(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("success"))
}
func writeFail(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(r.Header.Get("Content-Type") + ";" + r.Method + ": " + r.URL.Path))
}

type response struct {
	StatusCode int
	Body       []byte
}

func responseFromWriter(m *mockWriter) response {
	return response{
		StatusCode: m.statusCode,
		Body:       m.body,
	}
}

var successResponse = response{
	StatusCode: http.StatusOK,
	Body:       []byte("success"),
}

func TestRouter(t *testing.T) {

	tests := []struct {
		name    string
		reqFunc func() *http.Request
		init    func(router *Router)
		want    response
	}{{
		name: "GIVEN plain request THEN respond with matching route",
		reqFunc: func() *http.Request {
			req, _ := http.NewRequest(http.MethodGet, "/test/this/feat", nil)
			req.Header.Set("Content-Type", "application/json")
			return req
		},
		init: func(router *Router) {
			router.HandleMethod(http.MethodGet, "/test/", writeFail)
			router.HandleMethod(http.MethodGet, "/test/this", writeFail)
			router.HandleMethod(http.MethodGet, "/test/this/feat/out", writeFail)
			router.HandleMethod(http.MethodGet, "/test/this/feat2", writeFail)
			router.Handle("application/vnd.api+json", http.MethodGet, "/test/this/feat", writeFail)
			router.Handle("application/json", http.MethodPost, "/test/this/feat", writeSuccess)

			router.Handle("application/json", http.MethodGet, "/test/this/feat", writeSuccess)
		},
		want: successResponse,
	}, {
		name: "GIVEN request with param THEN respond with matching route",
		reqFunc: func() *http.Request {
			req, _ := http.NewRequest(http.MethodGet, "/domain/item/1", nil)
			req.Header.Set("Content-Type", "application/json")
			return req
		},
		init: func(router *Router) {
			router.HandleMethod(http.MethodGet, "/domain/", writeFail)
			router.HandleMethod(http.MethodGet, "/domain/item", writeFail)
			router.HandleMethod(http.MethodGet, "/domain/item/:id", writeSuccess)
			router.HandleMethod(http.MethodGet, "/domain/item/other", writeFail)

		},
		want: successResponse,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wr := &mockWriter{}
			r := NewRouter()
			tt.init(r)
			req := tt.reqFunc()
			r.ServeHTTP(wr, req)

			if !reflect.DeepEqual(tt.want, responseFromWriter(wr)) {
				t.Errorf("ServeHTTP() mismatch:\nwant:%v, got:%v", tt.want, responseFromWriter(wr))
			}
		})
	}
}
