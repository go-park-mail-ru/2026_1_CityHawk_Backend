package metrics

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const unknownRoute = "unknown"

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cityhawk_requests_total",
			Help: "Total number of requests handled by CityHawk services.",
		},
		[]string{"service", "protocol", "method", "route", "status"},
	)
	requestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cityhawk_request_errors_total",
			Help: "Total number of failed requests handled by CityHawk services.",
		},
		[]string{"service", "protocol", "method", "route", "status"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cityhawk_request_duration_seconds",
			Help:    "Request duration in seconds for CityHawk services.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "protocol", "method", "route", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal, requestErrorsTotal, requestDuration)
}

func Handler() http.Handler {
	return promhttp.Handler()
}

func HTTPMiddleware(service string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			statusValue := strconv.Itoa(recorder.status)
			route := httpRoute(r)
			requestsTotal.WithLabelValues(service, "http", r.Method, route, statusValue).Inc()
			if recorder.status >= http.StatusBadRequest {
				requestErrorsTotal.WithLabelValues(service, "http", r.Method, route, statusValue).Inc()
			}
			requestDuration.WithLabelValues(service, "http", r.Method, route, statusValue).Observe(time.Since(start).Seconds())
		})
	}
}

func UnaryServerInterceptor(service string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		code := status.Code(err)
		statusValue := code.String()
		method := grpcMethod(info.FullMethod)
		requestsTotal.WithLabelValues(service, "grpc", method, info.FullMethod, statusValue).Inc()
		if err != nil {
			requestErrorsTotal.WithLabelValues(service, "grpc", method, info.FullMethod, statusValue).Inc()
		}
		requestDuration.WithLabelValues(service, "grpc", method, info.FullMethod, statusValue).Observe(time.Since(start).Seconds())

		return resp, err
	}
}

func httpRoute(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	if r.URL == nil || r.URL.Path == "" {
		return unknownRoute
	}
	return r.URL.Path
}

func grpcMethod(fullMethod string) string {
	fullMethod = strings.Trim(fullMethod, "/")
	if fullMethod == "" {
		return unknownRoute
	}
	parts := strings.Split(fullMethod, "/")
	return parts[len(parts)-1]
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
