package main

import (
	"log"
	"sync"
	"time"
)

type Metrics struct {
	mu      sync.Mutex
	metrics map[string]map[string]struct {
		totalCalls   int
		totalLatency time.Duration
	}
}
type TxnMetric struct {
	Mutex            sync.RWMutex
	TotalElapsedTime time.Duration // Total time across all transactions
	TransactionCount int           // Total number of transactions processed
}

var TnxSum = TxnMetric{
	Mutex:            sync.RWMutex{},
	TotalElapsedTime: 0,
	TransactionCount: 0,
}

// GetMetricsInstance returns the singleton instance.
var metricsInstance *Metrics
var once sync.Once

func GetMetricsInstance() *Metrics {
	once.Do(func() {
		metricsInstance = &Metrics{
			metrics: make(map[string]map[string]struct {
				totalCalls   int
				totalLatency time.Duration
			}),
		}
	})
	return metricsInstance
}

// RecordCall stores metrics for a specific handler on a given server.
func (m *Metrics) RecordCall(serverID, handlerName string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.metrics[serverID]; !ok {
		m.metrics[serverID] = make(map[string]struct {
			totalCalls   int
			totalLatency time.Duration
		})
	}

	// Update the metrics for the handler on the given server.
	handlerMetrics := m.metrics[serverID][handlerName]
	handlerMetrics.totalCalls++
	handlerMetrics.totalLatency += latency
	m.metrics[serverID][handlerName] = handlerMetrics
}

// GetAllMetrics returns the complete metrics map.
func (m *Metrics) GetAllMetrics() map[string]map[string]struct {
	totalCalls   int
	totalLatency time.Duration
} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.metrics
}
func PrintMetricsForAllServersAndHandlers(client string) {
	metrics := GetMetricsInstance()
	allMetrics := metrics.GetAllMetrics()

	for serverID, handlers := range allMetrics {
		if (serverID == client) || client == "all" {
			log.Printf("Performance for Server: %s\t", serverID)
			for handlerName, metric := range handlers {
				avgLatency := time.Duration(0)
				if metric.totalCalls > 0 {
					avgLatency = metric.totalLatency / time.Duration(metric.totalCalls)
				}
				log.Printf("  Handler: %s    Total Calls: %d   Total Latency: %v    Average Latency: %v\n",
					handlerName, metric.totalCalls, metric.totalLatency, avgLatency)
			}
		}
	}
}
