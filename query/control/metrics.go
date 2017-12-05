package control

import "github.com/prometheus/client_golang/prometheus"

var queueingGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ifql_control_current_queueing",
	Help: "Number of queries currently queueing",
})
var requeueingGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ifql_control_current_requeueing",
	Help: "Number of queries currently requeueing",
})
var planningGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ifql_control_current_planning",
	Help: "Number of queries currently planning",
})
var executingGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "ifql_control_current_executing",
	Help: "Number of queries currently executing",
})

var queueingHist = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "ifql_control_queueing",
	Help:    "Histogram of queueing durations",
	Buckets: prometheus.ExponentialBuckets(1e-3, 5, 7),
})
var requeueingHist = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "ifql_control_requeueing",
	Help:    "Histogram of requeueing durations",
	Buckets: prometheus.ExponentialBuckets(1e-3, 5, 7),
})
var planningHist = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "ifql_control_planning",
	Help:    "Histogram of planning durations",
	Buckets: prometheus.ExponentialBuckets(1e-5, 5, 7),
})
var executingHist = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "ifql_control_executing",
	Help:    "Histogram of executing durations",
	Buckets: prometheus.ExponentialBuckets(1e-3, 5, 7),
})

func init() {
	prometheus.MustRegister(queueingGauge)
	prometheus.MustRegister(requeueingGauge)
	prometheus.MustRegister(planningGauge)
	prometheus.MustRegister(executingGauge)

	prometheus.MustRegister(queueingHist)
	prometheus.MustRegister(requeueingHist)
	prometheus.MustRegister(planningHist)
	prometheus.MustRegister(executingHist)
}
