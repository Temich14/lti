package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type LTIMetrics struct {
	launchDuration prometheus.Histogram
	launchSuccess  prometheus.Counter
	launchError    prometheus.Counter

	jwksDuration prometheus.Histogram
	jwksError    prometheus.Counter

	nrpsDuration     prometheus.Histogram
	nrpsError        prometheus.Counter
	agsRequestsTotal prometheus.Counter
	agsRequestErrors prometheus.Counter
}

func NewLTIMetrics() *LTIMetrics {

	m := &LTIMetrics{

		launchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "lti_launch_duration_seconds",
			Help:    "Duration of LTI launch",
			Buckets: prometheus.DefBuckets,
		}),

		launchSuccess: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lti_launch_success_total",
			Help: "Total successful launches",
		}),

		launchError: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lti_launch_error_total",
			Help: "Total failed launches",
		}),

		jwksDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "lti_jwks_fetch_duration_seconds",
			Help: "JWKS fetch duration",
		}),

		jwksError: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lti_jwks_fetch_error_total",
			Help: "JWKS fetch errors",
		}),

		nrpsDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "lti_nrps_request_duration_seconds",
			Help: "NRPS request duration",
		}),

		nrpsError: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "lti_nrps_request_error_total",
			Help: "NRPS request errors",
		}),

		agsRequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "lti_ags_requests_total",
				Help: "Total AGS requests",
			}),
		agsRequestErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "lti_ags_errors_total",
				Help: "Total AGS errors",
			}),
	}

	return m
}

func (m *LTIMetrics) Register(registerer prometheus.Registerer) error {
	if err := registerer.Register(m.launchDuration); err != nil {
		return err
	}
	if err := registerer.Register(m.launchSuccess); err != nil {
		return err
	}
	if err := registerer.Register(m.launchError); err != nil {
		return err
	}
	if err := registerer.Register(m.jwksDuration); err != nil {
		return err
	}
	if err := registerer.Register(m.jwksError); err != nil {
		return err
	}
	if err := registerer.Register(m.nrpsDuration); err != nil {
		return err
	}
	if err := registerer.Register(m.nrpsError); err != nil {
		return err
	}
	if err := registerer.Register(m.agsRequestsTotal); err != nil {
		return err
	}
	if err := registerer.Register(m.agsRequestErrors); err != nil {
		return err
	}
	return nil
}

func (m *LTIMetrics) IncAGSRequestsTotal() {
	m.agsRequestsTotal.Inc()
}

func (m *LTIMetrics) IncAGSRequestErrors() {
	m.agsRequestErrors.Inc()
}

func (m *LTIMetrics) ObserveLaunchDuration(d time.Duration) {
	m.launchDuration.Observe(d.Seconds())
}

func (m *LTIMetrics) IncLaunchSuccess() {
	m.launchSuccess.Inc()
}

func (m *LTIMetrics) IncLaunchError() {
	m.launchError.Inc()
}

func (m *LTIMetrics) ObserveJWKSFetch(d time.Duration) {
	m.jwksDuration.Observe(d.Seconds())
}

func (m *LTIMetrics) IncJWKSFetchError() {
	m.jwksError.Inc()
}

func (m *LTIMetrics) ObserveNRPSRequest(d time.Duration) {
	m.nrpsDuration.Observe(d.Seconds())
}

func (m *LTIMetrics) IncNRPSError() {
	m.nrpsError.Inc()
}
