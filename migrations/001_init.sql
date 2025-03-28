-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create rides table
CREATE TABLE rides (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    driver_id INTEGER REFERENCES users(id),
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    from_location POINT NOT NULL,
    to_location POINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'requested',
    price DECIMAL(10,2),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    finished_at TIMESTAMP WITH TIME ZONE
);

-- Create index for location-based queries
CREATE INDEX rides_from_location_idx ON rides USING GIST (from_location);
CREATE INDEX rides_to_location_idx ON rides USING GIST (to_location);

-- Create driver_locations table for real-time tracking
CREATE TABLE driver_locations (
    driver_id INTEGER PRIMARY KEY REFERENCES users(id),
    location POINT NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX driver_locations_idx ON driver_locations USING GIST (location); 