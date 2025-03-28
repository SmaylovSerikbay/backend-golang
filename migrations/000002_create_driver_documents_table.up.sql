CREATE TABLE IF NOT EXISTS driver_documents (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    car_brand VARCHAR(100) NOT NULL,
    car_model VARCHAR(100) NOT NULL,
    car_year VARCHAR(4) NOT NULL,
    car_color VARCHAR(50) NOT NULL,
    car_number VARCHAR(20) NOT NULL,
    driver_license_front VARCHAR(255) NOT NULL,
    driver_license_back VARCHAR(255) NOT NULL,
    car_registration_front VARCHAR(255) NOT NULL,
    car_registration_back VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
); 