-- Обновляем таблицу optimized_routes
ALTER TABLE optimized_routes
    ALTER COLUMN ride_id SET NOT NULL,
    ALTER COLUMN distance SET NOT NULL,
    ALTER COLUMN duration SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL,
    ALTER COLUMN distance TYPE DECIMAL(10,2);

-- Обновляем таблицу route_points
ALTER TABLE route_points
    ALTER COLUMN order_num SET NOT NULL,
    ALTER COLUMN ride_id SET NOT NULL,
    ALTER COLUMN route_id SET NOT NULL,
    ALTER COLUMN type SET NOT NULL,
    ALTER COLUMN address SET NOT NULL,
    ALTER COLUMN latitude SET NOT NULL,
    ALTER COLUMN longitude SET NOT NULL,
    ALTER COLUMN time SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL,
    ALTER COLUMN type TYPE VARCHAR(20),
    ALTER COLUMN address TYPE TEXT,
    ALTER COLUMN latitude TYPE DECIMAL(10,6),
    ALTER COLUMN longitude TYPE DECIMAL(10,6);

-- Добавляем индексы
CREATE INDEX IF NOT EXISTS idx_route_points_order_num ON route_points(order_num);
CREATE INDEX IF NOT EXISTS idx_route_points_ride_id ON route_points(ride_id);
CREATE INDEX IF NOT EXISTS idx_route_points_route_id ON route_points(route_id);
CREATE INDEX IF NOT EXISTS idx_route_points_booking_id ON route_points(booking_id);

-- Добавляем внешние ключи с каскадным удалением
ALTER TABLE route_points
    DROP CONSTRAINT IF EXISTS fk_route_points_route,
    ADD CONSTRAINT fk_route_points_route
        FOREIGN KEY (route_id)
        REFERENCES optimized_routes(id)
        ON DELETE CASCADE;

ALTER TABLE route_points
    DROP CONSTRAINT IF EXISTS fk_route_points_ride,
    ADD CONSTRAINT fk_route_points_ride
        FOREIGN KEY (ride_id)
        REFERENCES rides(id)
        ON DELETE CASCADE;

ALTER TABLE route_points
    DROP CONSTRAINT IF EXISTS fk_route_points_booking,
    ADD CONSTRAINT fk_route_points_booking
        FOREIGN KEY (booking_id)
        REFERENCES bookings(id)
        ON DELETE SET NULL; 