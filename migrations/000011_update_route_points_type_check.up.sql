-- Удаляем старое ограничение
ALTER TABLE route_points DROP CONSTRAINT route_points_type_check;

-- Добавляем новое ограничение с типом 'start'
ALTER TABLE route_points ADD CONSTRAINT route_points_type_check 
    CHECK (type IN ('pickup', 'dropoff', 'start'));

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
    ALTER COLUMN time SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL;

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