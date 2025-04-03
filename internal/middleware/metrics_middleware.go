package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal - общее количество запросов
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP запросов",
		},
		[]string{"method", "endpoint", "status"},
	)

	// RequestDuration - длительность запросов
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP запросов в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// RequestsInFlight - количество запросов в обработке
	RequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Текущее количество запросов в обработке",
		},
	)

	// DGISRequestsTotal - общее количество запросов к 2ГИС API
	DGISRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dgis_requests_total",
			Help: "Общее количество запросов к 2ГИС API",
		},
		[]string{"endpoint", "status", "cached"},
	)

	// DGISRequestDuration - длительность запросов к 2ГИС API
	DGISRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dgis_request_duration_seconds",
			Help:    "Длительность запросов к 2ГИС API в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "cached"},
	)
)

// PrometheusMiddleware собирает метрики для HTTP запросов
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Увеличиваем счетчик запросов в обработке
		RequestsInFlight.Inc()
		defer RequestsInFlight.Dec()

		// Фиксируем время начала запроса
		start := time.Now()

		// Обрабатываем запрос
		c.Next()

		// Вычисляем длительность запроса
		duration := time.Since(start).Seconds()

		// Получаем статус код и эндпоинт
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		// Увеличиваем счетчик запросов
		RequestsTotal.WithLabelValues(c.Request.Method, endpoint, status).Inc()

		// Добавляем длительность запроса
		RequestDuration.WithLabelValues(c.Request.Method, endpoint).Observe(duration)
	}
}

// TrackDGISRequest отслеживает запрос к 2ГИС API
func TrackDGISRequest(endpoint string, status string, cached bool, duration time.Duration) {
	// Увеличиваем счетчик запросов к 2ГИС API
	cachedStr := strconv.FormatBool(cached)
	DGISRequestsTotal.WithLabelValues(endpoint, status, cachedStr).Inc()

	// Добавляем длительность запроса к 2ГИС API
	DGISRequestDuration.WithLabelValues(endpoint, cachedStr).Observe(duration.Seconds())
}
