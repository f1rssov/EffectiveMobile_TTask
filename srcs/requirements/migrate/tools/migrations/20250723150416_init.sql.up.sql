CREATE TABLE IF NOT EXISTS subscriptions (
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,       
    end_date DATE,               
    PRIMARY KEY (user_id, service_name, start_date)
);