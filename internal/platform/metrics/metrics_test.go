package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/grpc"
)

func TestHTTPMiddlewareRecordsRequests(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	req.SetPathValue("id", "123")
	rec := httptest.NewRecorder()

	HTTPMiddleware("test-service")(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d", rec.Code)
	}
	if httpRoute(req) != "/api/test" {
		t.Fatalf("httpRoute() = %q", httpRoute(req))
	}
	if Handler() == nil {
		t.Fatal("Handler() = nil")
	}
}

func TestUnaryServerInterceptorRecordsRequests(t *testing.T) {
	interceptor := UnaryServerInterceptor("test-service")
	resp, err := interceptor(context.Background(), "request", &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}, func(_ context.Context, req any) (any, error) {
		return "response", nil
	})
	if err != nil || resp != "response" {
		t.Fatalf("interceptor() = (%v, %v)", resp, err)
	}
	if grpcMethod("/pkg.Service/Method") != "Method" {
		t.Fatalf("grpcMethod() mismatch")
	}
	if grpcMethod("") != unknownRoute {
		t.Fatalf("grpcMethod(empty) mismatch")
	}
}
