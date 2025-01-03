package services

import "github.com/prometheus/client_golang/prometheus"

type Monitoring interface {
	Record(event Event)
	GetRegistry() *prometheus.Registry
}

type EventName uint

type Event struct {
	id     EventName
	params []interface{}
}

func (e Event) GetID() EventName {
	return e.id
}
