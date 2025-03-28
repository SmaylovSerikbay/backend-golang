-- Удаляем старые таблицы, если они существуют
DROP TABLE IF EXISTS route_points CASCADE;
DROP TABLE IF EXISTS optimized_routes CASCADE;

-- Создаем таблицу для оптимизированных маршрутов заново
CREATE TABLE optimized_routes (
    id SERIAL PRIMARY KEY,
    ride_id INTEGER NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    distance DECIMAL(10, 2) NOT NULL DEFAULT 0,
    duration INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT optimized_routes_ride_id_unique UNIQUE(ride_id)
);

-- Создаем таблицу для точек маршрута заново
CREATE TABLE route_points (
    id SERIAL PRIMARY KEY,
    order_num INTEGER NOT NULL,
    ride_id INTEGER NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    route_id INTEGER NOT NULL REFERENCES optimized_routes(id) ON DELETE CASCADE,
    booking_id INTEGER REFERENCES bookings(id) ON DELETE SET NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('pickup', 'dropoff', 'start')),
    address TEXT NOT NULL,
    latitude DECIMAL(10,6) NOT NULL,
    longitude DECIMAL(10,6) NOT NULL,
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Сбрасываем последовательности
ALTER SEQUENCE route_points_id_seq RESTART WITH 1;
ALTER SEQUENCE optimized_routes_id_seq RESTART WITH 1;

-- Создаем индексы для оптимизации запросов
CREATE INDEX idx_route_points_order_num ON route_points(order_num);
CREATE INDEX idx_route_points_route_id ON route_points(route_id);
CREATE INDEX idx_route_points_ride_id ON route_points(ride_id);
CREATE INDEX idx_route_points_booking_id ON route_points(booking_id); 