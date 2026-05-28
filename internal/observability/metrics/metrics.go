// Package metrics provides Prometheus metrics and HTTP/database instrumentation.
package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	namespace          = "go_tweets"
	unknownLabelValue  = "unknown"
	queryDurationName  = "db_query_duration_seconds"
	requestsTotalName  = "http_requests_total"
	requestErrorsName  = "http_request_errors_total"
	requestLatencyName = "http_request_duration_seconds"
)

type Registry struct {
	registry            *prometheus.Registry
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestErrors   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	dbQueryDuration     *prometheus.HistogramVec
}

func NewRegistry() *Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	httpRequestsTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      requestsTotalName,
		Help:      "Total number of HTTP requests by route, method and status code.",
	}, []string{"route", "method", "status_code"})

	httpRequestErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      requestErrorsName,
		Help:      "Total number of HTTP error responses by route, method and status code.",
	}, []string{"route", "method", "status_code"})

	httpRequestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      requestLatencyName,
		Help:      "HTTP request latency in seconds by route, method and status code.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"route", "method", "status_code"})

	dbQueryDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      queryDurationName,
		Help:      "Database query duration in seconds by operation and table.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"operation", "table"})

	registry.MustRegister(
		httpRequestsTotal,
		httpRequestErrors,
		httpRequestDuration,
		dbQueryDuration,
	)

	return &Registry{
		registry:            registry,
		httpRequestsTotal:   httpRequestsTotal,
		httpRequestErrors:   httpRequestErrors,
		httpRequestDuration: httpRequestDuration,
		dbQueryDuration:     dbQueryDuration,
	}
}

func (r *Registry) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		statusCode := strconv.Itoa(c.Writer.Status())
		r.httpRequestsTotal.WithLabelValues(route, c.Request.Method, statusCode).Inc()
		r.httpRequestDuration.WithLabelValues(route, c.Request.Method, statusCode).Observe(time.Since(startedAt).Seconds())

		if c.Writer.Status() >= http.StatusBadRequest {
			r.httpRequestErrors.WithLabelValues(route, c.Request.Method, statusCode).Inc()
		}
	}
}

func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

func (r *Registry) ObserveQuery(query string, duration time.Duration) {
	operation, table := queryLabels(query)
	r.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

func queryLabels(query string) (string, string) {
	normalized := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	if normalized == "" {
		return unknownLabelValue, unknownLabelValue
	}

	switch {
	case strings.HasPrefix(normalized, "select"):
		return "select", tableAfterKeyword(normalized, " from ")
	case strings.HasPrefix(normalized, "insert"):
		return "insert", tableAfterKeyword(normalized, " into ")
	case strings.HasPrefix(normalized, "update"):
		return "update", tableAfterPrefix(normalized, "update ")
	case strings.HasPrefix(normalized, "delete"):
		return "delete", tableAfterKeyword(normalized, " from ")
	default:
		return unknownLabelValue, unknownLabelValue
	}
}

func tableAfterKeyword(query string, keyword string) string {
	index := strings.Index(query, keyword)
	if index == -1 {
		return unknownLabelValue
	}

	return cleanTableName(query[index+len(keyword):])
}

func tableAfterPrefix(query string, prefix string) string {
	if !strings.HasPrefix(query, prefix) {
		return unknownLabelValue
	}

	return cleanTableName(strings.TrimPrefix(query, prefix))
}

func cleanTableName(fragment string) string {
	if fragment == "" {
		return unknownLabelValue
	}

	fields := strings.Fields(fragment)
	if len(fields) == 0 {
		return unknownLabelValue
	}

	table := strings.Trim(fields[0], "` ,;")
	if table == "" {
		return unknownLabelValue
	}

	return table
}
