package monitoring

import (
	"log"
	"net/http"
	"os"
	"time"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsService struct {
	Counter   map[string]*prometheus.CounterVec
	Gauge     map[string]prometheus.Gauge
	Summary   map[string]prometheus.Summary
	Histogram map[string]*prometheus.HistogramVec
}

func NewMetricsService(router *httprouter.Router) (*MetricsService, error) {
	mservice := MetricsService{
		Counter:   make(map[string]*prometheus.CounterVec),
		Gauge:     make(map[string]prometheus.Gauge),
		Summary:   make(map[string]prometheus.Summary),
		Histogram: make(map[string]*prometheus.HistogramVec),
	}

	mservice.addDeviceMetrics()
	err := mservice.RunPrometheus(router)

	logger.LogDebug("start server metrics...", err)

	return &mservice, nil
}

func (m *MetricsService) addDeviceMetrics() {

	createAddDevice := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "go_metrics_devices",
		Name:      "bootstrap_devices",
		Help:      "Общее количество пришедших на сервер устройств",
	}, []string{"type"})
	m.Counter["bootstrap_devices"] = createAddDevice

	workingDevCreate := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "go_metrics_devices",
			Name:      "active_devices",
			Help:      "Количество активных подключенных устройств",
		})
	m.Gauge["active_devices"] = workingDevCreate

	requestProcessingTimeHistogramMs := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "go_metrics_devices",
			Name:      "request_processing_time_histogram_ms",
			Help:      "Продолжительность исполнения запроса Histogram",
			Buckets:   []float64{0.1, 0.15, 0.2, 0.25, 0.3},
		}, []string{"status", "method"})
	m.Histogram["request_processing_time_histogram_ms"] = requestProcessingTimeHistogramMs

	requestProcessingTimeSummaryMs := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: "example_go_metrics_orders",
			Name:      "request_processing_time_summary_ms",
			Help:      "Продолжительность исполнения запроса Summary",

			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		})
	m.Summary["request_processing_time_summary_ms"] = requestProcessingTimeSummaryMs

}

func (m *MetricsService) RunPrometheus(router *httprouter.Router) error {

	var err error
	if len(m.Counter) > 0 {
		for _, counter := range m.Counter {
			err = prometheus.Register(counter)
			if err != nil {
				return err
			}
		}
	}

	if len(m.Gauge) > 0 {
		for _, gauge := range m.Gauge {
			err = prometheus.Register(gauge)
			if err != nil {
				return err
			}
		}
	}

	if len(m.Summary) > 0 {
		for _, summary := range m.Summary {
			err = prometheus.Register(summary)
			if err != nil {
				return err
			}
		}
	}

	if len(m.Histogram) > 0 {
		for _, histogram := range m.Histogram {
			err = prometheus.Register(histogram)
			if err != nil {
				return err
			}
		}
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(":"+os.Getenv("METRICS_PORT"), nil)
		log.Println(err)
	}()

	return nil
}

func (m *MetricsService) AddDevInst() {
	m.Counter["bootstrap_devices"].With(prometheus.Labels{"type": "bootstrap_devices"}).Inc()
}

func (m *MetricsService) AddActiveInst() time.Time {
	logger.LogDebug("Metrics add active devices")

	m.Gauge["active_devices"].Inc()
	return time.Now()
}

func (m *MetricsService) MetricsResultReq(now time.Time, err string) {
	logger.LogDebug("Metrics result request")

	m.Histogram["request_processing_time_histogram_ms"].With(prometheus.Labels{"method": "CreateOrderV1", "status": err}).Observe(time.Since(now).Seconds())
	m.Summary["request_processing_time_summary_ms"].Observe(time.Since(now).Seconds())
	m.Gauge["active_devices"].Dec()
}
