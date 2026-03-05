package metrics

import (
	"net/http"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	_ "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Metrics struct {
	logger         *log.Logger
	ValidQueries   prometheus.Counter
	InvalidQueries prometheus.Counter
}

func NewMetrics(logger *log.Logger) (*Metrics, *http.Handler) {

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(
				collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/sched/latencies:seconds")},
			),
		),
	)

	validQ := promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "birthday_registered_valid_total",
		Help: "Total number of valid birthdays registered",
	})
	invalidQ := promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "birthday_invalid_total",
		Help: "Total number of invalid birthdays attempted to register",
	})
	handler := promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	)

	return &Metrics{
		logger:         logger,
		ValidQueries:   validQ,
		InvalidQueries: invalidQ,
	}, &handler
}
