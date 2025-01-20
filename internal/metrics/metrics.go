package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gitlab.ozon.dev/ashadkhamov/homework/internal/interfaces"
)

type metrics struct {
	ordersServed prometheus.Counter
}

var once sync.Once
var instance interfaces.Metrics

func GetMetrics() interfaces.Metrics {
	once.Do(func() {
		instance = &metrics{
			ordersServed: prometheus.NewCounter(prometheus.CounterOpts{
				Name: "orders_served_total",
				Help: "Total number of orders served",
			}),
		}

		prometheus.MustRegister(instance.(*metrics).ordersServed)
	})
	return instance
}

func (m *metrics) IncOrdersServed() {
	m.ordersServed.Inc()
}
