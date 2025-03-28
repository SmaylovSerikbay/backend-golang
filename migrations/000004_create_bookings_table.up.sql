CREATE TABLE IF NOT EXISTS "bookings" (
    "id" SERIAL PRIMARY KEY,
    "ride_id" INTEGER NOT NULL,
    "passenger_id" INTEGER NOT NULL,
    "pickup_address" TEXT NOT NULL,
    "dropoff_address" TEXT NOT NULL,
    "pickup_location" POINT NOT NULL,
    "dropoff_location" POINT NOT NULL,
    "seats_count" INTEGER NOT NULL,
    "status" VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'started', 'rejected', 'cancelled', 'completed')),
    "booking_type" VARCHAR(20) NOT NULL DEFAULT 'regular',
    "price" DECIMAL(10, 2) NOT NULL,
    "comment" TEXT DEFAULT '',
    "reject_reason" TEXT DEFAULT '',
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    FOREIGN KEY ("ride_id") REFERENCES "rides" ("id"),
    FOREIGN KEY ("passenger_id") REFERENCES "users" ("id")
);

-- Создаем индексы для ускорения запросов
CREATE INDEX IF NOT EXISTS "idx_bookings_ride_id" ON "bookings" ("ride_id");
CREATE INDEX IF NOT EXISTS "idx_bookings_passenger_id" ON "bookings" ("passenger_id");
CREATE INDEX IF NOT EXISTS "idx_bookings_status" ON "bookings" ("status"); 