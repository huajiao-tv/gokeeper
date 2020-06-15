package metrics

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ServerMetrics struct {
	NodeEventCounter *prometheus.CounterVec
	NodeAliveCounter *prometheus.CounterVec
}

var (
	Metrics *ServerMetrics
)

func Init(promListen string) {
	Metrics = NewServerMetrics()
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(promListen, nil)
		if err != nil {
			log.Println("prometheus ListenAndServe error:", err)
		}
	}()
}

func NewServerMetrics() *ServerMetrics {
	serverMetrics := &ServerMetrics{
		NodeEventCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "keeper_node_event_total",
				Help: "keeper node event",
			},
			[]string{"type", "event", "id", "domain", "hostname"},
		),
		NodeAliveCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "keeper_node_alive_total",
				Help: "keeper node alive num",
			},
			[]string{"id", "domain", "hostname"},
		),
	}
	prometheus.MustRegister(serverMetrics.NodeEventCounter)
	prometheus.MustRegister(serverMetrics.NodeAliveCounter)
	return serverMetrics
}

//  节点事件添加
func (m *ServerMetrics) AddNodeEvent(typ, event, id, domain, hostname string, v int64) {
	labels := map[string]string{
		"type":     typ,
		"event":    event,
		"id":       id,
		"domain":   domain,
		"hostname": hostname,
	}
	m.NodeEventCounter.With(labels).Add(float64(v))
}

// 存活的节点添加
func (m *ServerMetrics) AddNodeAlive(id, domain, hostname string, v uint64) {
	labels := map[string]string{
		"id":       id,
		"domain":   domain,
		"hostname": hostname,
	}
	m.NodeAliveCounter.With(labels).Add(float64(v))
}
