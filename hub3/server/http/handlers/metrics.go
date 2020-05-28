package handlers

import (
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	//"github.com/thoas/stats"
)

var (
	dflBuckets = []float64{300, 1200, 5000}
)

const (
	reqsName    = "requests_total"
	latencyName = "request_duration_milliseconds"
)

func RegisterMetrics(r chi.Router) {
	//stats := stats.New()
	//r.Use(StatsMiddleware(stats))

	//// stats page
	//r.Get("/api/stats/http", func(w http.ResponseWriter, r *http.Request) {
	//stats := stats.Data()
	//render.JSON(w, r, stats)
	//return
	//})

	// r.Handle("/metrics", prometheus.Handler())

}

// PrometheusMiddleware is a handler that exposes prometheus metrics for the number of requests,
// the latency and the response size, partitioned by status code, method and HTTP path.
type prometheusMiddleware struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// NewMiddleware returns a new prometheus Middleware handler.
func NewPrometheuusMiddleware(name string, buckets ...float64) *prometheusMiddleware {
	var m prometheusMiddleware
	m.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.reqs)

	if len(buckets) == 0 {
		buckets = dflBuckets
	}
	m.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": name},
		Buckets:     buckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.latency)
	return &m
}

//func (m *prometheusMiddleware) Handler() func(http.Handler) http.Handler {
//return func(next http.Handler) http.Handler {
//fn := func(w http.ResponseWriter, r *http.Request) {
//start := time.Now()
//next.ServeHTTP(w, r)
//m.reqs.WithLabelValues(http.StatusText(w.Header()., r.Method, r.URL.Path).Inc()
//m.latency.WithLabelValues(http.StatusText(res.Status()), r.Method, r.URL.Path).Observe(float64(time.Since(start).Nanoseconds()) / 1000000)
//}
//return http.HandlerFunc(fn)
//}
//}

//func StatsMiddleware(middleware *stats.Stats) func(http.Handler) http.Handler {
//return func(next http.Handler) http.Handler {
//fn := func(w http.ResponseWriter, r *http.Request) {
//beginning, recorder := middleware.Begin(w)
//next.ServeHTTP(w, r)
//middleware.End(beginning, stats.WithRecorder(recorder))
//}
//return http.HandlerFunc(fn)
//}
//}
