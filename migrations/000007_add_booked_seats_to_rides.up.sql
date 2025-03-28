ALTER TABLE rides ADD COLUMN booked_seats INTEGER NOT NULL DEFAULT 0;

-- Обновляем существующие записи, подсчитывая количество подтвержденных бронирований
UPDATE rides r 
SET booked_seats = (
    SELECT COALESCE(SUM(b.seats_count), 0)
    FROM bookings b
    WHERE b.ride_id = r.id 
    AND b.status = 'approved'
); 