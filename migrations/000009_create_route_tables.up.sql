-- Создаем таблицу для оптимизированных маршрутов
CREATE TABLE IF NOT EXISTS optimized_routes (
    id SERIAL PRIMARY KEY,
    ride_id INTEGER NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    distance DECIMAL(10, 2) NOT NULL,
    duration INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ride_id)
);

-- Создаем таблицу для точек маршрута
CREATE TABLE IF NOT EXISTS route_points (
    id SERIAL PRIMARY KEY,
    order_num INTEGER NOT NULL,
    ride_id INTEGER NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    route_id INTEGER NOT NULL REFERENCES optimized_routes(id) ON DELETE CASCADE,
    booking_id INTEGER REFERENCES bookings(id) ON DELETE SET NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('pickup', 'dropoff')),
    address TEXT NOT NULL,
    location TEXT NOT NULL,
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_route_points_route_id ON route_points(route_id);
CREATE INDEX IF NOT EXISTS idx_route_points_ride_id ON route_points(ride_id);
CREATE INDEX IF NOT EXISTS idx_route_points_booking_id ON route_points(booking_id); 