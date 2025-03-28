-- Удаляем внешние ключи
ALTER TABLE route_points DROP CONSTRAINT IF EXISTS fk_route_points_route;
ALTER TABLE route_points DROP CONSTRAINT IF EXISTS fk_route_points_ride;
ALTER TABLE route_points DROP CONSTRAINT IF EXISTS fk_route_points_booking;

-- Удаляем индексы
DROP INDEX IF EXISTS idx_route_points_order_num;
DROP INDEX IF EXISTS idx_route_points_ride_id;
DROP INDEX IF EXISTS idx_route_points_route_id;
DROP INDEX IF EXISTS idx_route_points_booking_id;

-- Возвращаем типы данных
ALTER TABLE optimized_routes
    ALTER COLUMN ride_id DROP NOT NULL,
    ALTER COLUMN distance DROP NOT NULL,
    ALTER COLUMN duration DROP NOT NULL,
    ALTER COLUMN created_at DROP NOT NULL,
    ALTER COLUMN updated_at DROP NOT NULL,
    ALTER COLUMN distance TYPE FLOAT;

ALTER TABLE route_points
    ALTER COLUMN order_num DROP NOT NULL,
    ALTER COLUMN ride_id DROP NOT NULL,
    ALTER COLUMN route_id DROP NOT NULL,
    ALTER COLUMN type DROP NOT NULL,
    ALTER COLUMN address DROP NOT NULL,
    ALTER COLUMN latitude DROP NOT NULL,
    ALTER COLUMN longitude DROP NOT NULL,
    ALTER COLUMN time DROP NOT NULL,
    ALTER COLUMN created_at DROP NOT NULL,
    ALTER COLUMN updated_at DROP NOT NULL,
    ALTER COLUMN type TYPE VARCHAR,
    ALTER COLUMN address TYPE VARCHAR,
    ALTER COLUMN latitude TYPE FLOAT,
    ALTER COLUMN longitude TYPE FLOAT; 