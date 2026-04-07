CREATE TABLE IF NOT EXISTS availability_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('recurring', 'one-time')),
    day_of_week INT,
    date DATE,
    time_ranges JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS blocked_days (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_date DATE NOT NULL,
    slot_start_time TIME NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    recurrence VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (recurrence IN ('none', 'daily', 'weekly', 'yearly')),
    day_of_week INT,
    end_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
