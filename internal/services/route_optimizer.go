package services

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"taxi-backend/internal/models"
	"time"
)

type Location struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// ParseLocation парсит координаты из строки
func ParseLocation(location string) Location {
	location = strings.Trim(location, "()")
	parts := strings.Split(location, ",")
	if len(parts) != 2 {
		return Location{}
	}
	lng, _ := strconv.ParseFloat(parts[0], 64)
	lat, _ := strconv.ParseFloat(parts[1], 64)
	return Location{Latitude: lat, Longitude: lng}
}

type RouteOptimizer struct{}

func NewRouteOptimizer(apiKey string) *RouteOptimizer {
	return &RouteOptimizer{}
}

// OptimizeRoute оптимизирует маршрут для поездки
func (ro *RouteOptimizer) OptimizeRoute(ride *models.Ride, bookings []models.Booking) (*models.OptimizedRoute, error) {
	// Собираем точки маршрута
	var points []models.RoutePoint

	// Добавляем точки pickup только для подтвержденных бронирований
	orderNum := 1
	for _, booking := range bookings {
		// Пропускаем отклоненные и отмененные бронирования
		if booking.Status != "approved" {
			continue
		}

		pickupID := booking.ID
		// Парсим координаты из строки
		loc := ParseLocation(booking.PickupLocation)
		points = append(points, models.RoutePoint{
			RideID:    ride.ID,
			BookingID: &pickupID,
			Type:      "pickup",
			Address:   booking.PickupAddress,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
			Time:      time.Now(),
			OrderNum:  orderNum,
		})
		orderNum++
	}

	// Если нет подтвержденных бронирований, возвращаем пустой маршрут
	if len(points) == 0 {
		return &models.OptimizedRoute{
			RideID:   ride.ID,
			Points:   []models.RoutePoint{},
			Distance: 0,
			Duration: 0,
		}, nil
	}

	// Оптимизируем порядок точек pickup
	optimizedPoints, totalDistance, totalDuration := ro.findOptimalRoute(points)

	// Добавляем точки dropoff в конец без оптимизации
	var finalPoints []models.RoutePoint
	finalPoints = append(finalPoints, optimizedPoints...)

	// Добавляем точки dropoff только для подтвержденных бронирований
	for _, point := range optimizedPoints {
		if point.BookingID != nil {
			for _, booking := range bookings {
				if booking.ID == *point.BookingID && booking.Status == "approved" {
					// Парсим координаты из строки
					loc := ParseLocation(booking.DropoffLocation)
					dropoffPoint := models.RoutePoint{
						RideID:    ride.ID,
						BookingID: point.BookingID,
						Type:      "dropoff",
						Address:   booking.DropoffAddress,
						Latitude:  loc.Latitude,
						Longitude: loc.Longitude,
						Time:      time.Now(),
						OrderNum:  orderNum,
					}
					finalPoints = append(finalPoints, dropoffPoint)
					orderNum++
					break
				}
			}
		}
	}

	// Создаем оптимизированный маршрут
	route := &models.OptimizedRoute{
		RideID:   ride.ID,
		Points:   finalPoints,
		Distance: totalDistance,
		Duration: totalDuration,
	}

	return route, nil
}

// findOptimalRoute находит оптимальный порядок посещения точек
func (ro *RouteOptimizer) findOptimalRoute(points []models.RoutePoint) ([]models.RoutePoint, float64, int) {
	n := len(points)
	if n <= 1 {
		return points, 0, 0
	}

	// Находим конечную точку (Караганда)
	var destinationLat, destinationLng float64
	for _, point := range points {
		if point.Type == "dropoff" {
			destinationLat = point.Latitude
			destinationLng = point.Longitude
			break
		}
	}

	// Создаем матрицу расстояний и направлений для точек pickup
	distances := make([][]float64, n)
	durations := make([][]int, n)
	directionScores := make([]float64, n)

	// Рассчитываем направление для каждой точки относительно конечного города
	for i := range points {
		// Вычисляем угол между текущей точкой и конечным городом
		pointToDestAngle := math.Atan2(
			destinationLat-points[i].Latitude,
			destinationLng-points[i].Longitude,
		)
		// Нормализуем угол в радианах от -π до π
		if pointToDestAngle < 0 {
			pointToDestAngle += 2 * math.Pi
		}
		// Присваиваем score в зависимости от того, насколько точка "по пути"
		directionScores[i] = math.Cos(pointToDestAngle)

		distances[i] = make([]float64, n)
		durations[i] = make([]int, n)
		for j := range distances[i] {
			if i != j {
				dist, dur := ro.getDistanceAndDuration(points[i], points[j])
				// Корректируем расстояние с учетом направления движения
				directionBonus := (directionScores[j] + 1) / 2 // нормализуем от 0 до 1
				distances[i][j] = dist * (2 - directionBonus)  // уменьшаем вес расстояния для точек "по пути"
				durations[i][j] = dur
			}
		}
	}

	// Начинаем с точки водителя (индекс 0)
	visited := make([]bool, n)
	visited[0] = true
	path := []int{0}
	current := 0
	totalDistance := 0.0
	totalDuration := 0

	// Находим следующую точку с учетом направления
	for len(path) < n {
		minScore := float64(1e9)
		next := -1

		for i := 1; i < n; i++ {
			if !visited[i] {
				// Учитываем как расстояние, так и направление движения
				score := distances[current][i] * (2 - directionScores[i])
				if score < minScore {
					minScore = score
					next = i
				}
			}
		}

		if next != -1 {
			visited[next] = true
			path = append(path, next)
			totalDistance += distances[current][next]
			totalDuration += durations[current][next]
			current = next
		}
	}

	// Формируем оптимизированный маршрут
	optimizedPoints := make([]models.RoutePoint, n)
	for i, idx := range path {
		point := points[idx]
		point.OrderNum = i + 1
		optimizedPoints[i] = point
	}

	return optimizedPoints, totalDistance, totalDuration
}

// getDistanceAndDuration получает расстояние и время между двумя точками
func (ro *RouteOptimizer) getDistanceAndDuration(from, to models.RoutePoint) (float64, int) {
	// Используем OSRM для получения реального расстояния и времени
	url := fmt.Sprintf(
		"http://router.project-osrm.org/route/v1/driving/%f,%f;%f,%f?overview=false",
		from.Longitude, from.Latitude,
		to.Longitude, to.Latitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		// В случае ошибки используем приближенное расстояние
		distance, duration := ro.calculateApproximateDistance(
			Location{Latitude: from.Latitude, Longitude: from.Longitude},
			Location{Latitude: to.Latitude, Longitude: to.Longitude},
		)
		return distance, duration
	}
	defer resp.Body.Close()

	var result struct {
		Routes []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
		} `json:"routes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		distance, duration := ro.calculateApproximateDistance(
			Location{Latitude: from.Latitude, Longitude: from.Longitude},
			Location{Latitude: to.Latitude, Longitude: to.Longitude},
		)
		return distance, duration
	}

	if len(result.Routes) > 0 {
		return result.Routes[0].Distance, int(result.Routes[0].Duration)
	}

	distance, duration := ro.calculateApproximateDistance(
		Location{Latitude: from.Latitude, Longitude: from.Longitude},
		Location{Latitude: to.Latitude, Longitude: to.Longitude},
	)
	return distance, duration
}

// calculateApproximateDistance вычисляет приближенное расстояние между двумя точками
func (ro *RouteOptimizer) calculateApproximateDistance(from, to Location) (float64, int) {
	// Используем формулу гаверсинуса для вычисления расстояния между двумя точками
	R := 6371000.0 // Радиус Земли в метрах
	φ1 := from.Latitude * math.Pi / 180
	φ2 := to.Latitude * math.Pi / 180
	dφ := (to.Latitude - from.Latitude) * math.Pi / 180
	dλ := (to.Longitude - from.Longitude) * math.Pi / 180

	a := math.Sin(dφ/2)*math.Sin(dφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*math.Sin(dλ/2)*math.Sin(dλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance := R * c
	// Предполагаем среднюю скорость 40 км/ч
	duration := int(distance / (40 * 1000 / 3600)) // Конвертируем в секунды

	return distance, duration
}
