package monitoring

import (
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/models/entity"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	log "github.com/sirupsen/logrus"
)

type PrometheusAppMonitor struct {
	registry *prometheus.Registry
}

func NewPrometheusAppMonitor() *PrometheusAppMonitor {
	monitor := &PrometheusAppMonitor{registry: prometheus.NewRegistry()}
	prometheus.Unregister(collectors.NewGoCollector())
	prometheus.Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return monitor
}

func (p *PrometheusAppMonitor) GetRegistry() *prometheus.Registry {
	return p.registry
}

func (p *PrometheusAppMonitor) Record(event entity.Event) {
	switch event.GetID() {
	default:
		log.Errorf("prometheus app monitoring: invalid event id [%d]", event.GetID())
	}
}
