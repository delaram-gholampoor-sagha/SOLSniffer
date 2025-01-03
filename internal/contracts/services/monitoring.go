package services

import (
	"github.com/delaram-gholampoor-sagha/SOLSniffer/internal/models/entity"
	"github.com/prometheus/client_golang/prometheus"
)

type Monitoring interface {
	Record(event entity.Event)
	GetRegistry() *prometheus.Registry
}
