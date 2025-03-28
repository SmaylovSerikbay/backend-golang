CREATE TABLE IF NOT EXISTS rides (
    id SERIAL PRIMARY KEY,
    passenger_id INTEGER NULL DEFAULT NULL CHECK (passenger_id IS NULL OR passenger_id > 0) REFERENCES users(id),
    driver_id INTEGER NOT NULL CHECK (driver_id > 0) REFERENCES users(id),
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    from_location POINT NOT NULL,
    to_location POINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    seats_count INTEGER NOT NULL DEFAULT 4 CHECK (seats_count > 0),
    departure_date TIMESTAMP WITH TIME ZONE NOT NULL,
    comment TEXT,
    front_seat_price DECIMAL(10,2) CHECK (front_seat_price IS NULL OR front_seat_price >= 0),
    back_seat_price DECIMAL(10,2) CHECK (back_seat_price IS NULL OR back_seat_price >= 0),
    cancellation_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);