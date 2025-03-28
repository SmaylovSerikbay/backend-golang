-- Удаляем новое ограничение
ALTER TABLE route_points DROP CONSTRAINT route_points_type_check;

-- Восстанавливаем старое ограничение
ALTER TABLE route_points ADD CONSTRAINT route_points_type_check 
    CHECK (type IN ('pickup', 'dropoff')); 