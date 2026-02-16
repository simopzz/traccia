CREATE TABLE trips (
    id SERIAL PRIMARY KEY,
    user_id UUID,
    name TEXT NOT NULL,
    destination TEXT DEFAULT '',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    trip_id INTEGER NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    event_date DATE NOT NULL,
    title TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'activity',
    location TEXT DEFAULT '',
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    pinned BOOLEAN DEFAULT FALSE,
    position INTEGER NOT NULL DEFAULT 0,
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE flight_details (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL UNIQUE REFERENCES events(id) ON DELETE CASCADE,
    airline TEXT DEFAULT '',
    flight_number TEXT DEFAULT '',
    departure_airport TEXT DEFAULT '',
    arrival_airport TEXT DEFAULT '',
    departure_terminal TEXT DEFAULT '',
    arrival_terminal TEXT DEFAULT '',
    departure_gate TEXT DEFAULT '',
    arrival_gate TEXT DEFAULT '',
    booking_reference TEXT DEFAULT ''
);

CREATE TABLE lodging_details (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL UNIQUE REFERENCES events(id) ON DELETE CASCADE,
    check_in_time TIMESTAMPTZ,
    check_out_time TIMESTAMPTZ,
    booking_reference TEXT DEFAULT ''
);

CREATE TABLE transit_details (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL UNIQUE REFERENCES events(id) ON DELETE CASCADE,
    origin TEXT DEFAULT '',
    destination TEXT DEFAULT '',
    transport_mode TEXT DEFAULT ''
);

CREATE INDEX idx_trips_user_id ON trips(user_id);
CREATE INDEX idx_events_trip_date_pos ON events(trip_id, event_date, position);
