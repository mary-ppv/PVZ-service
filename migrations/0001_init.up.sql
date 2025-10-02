CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('employee', 'moderator'))
);

CREATE TABLE IF NOT EXISTS pvz (
    id UUID PRIMARY KEY,
    city TEXT NOT NULL CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань')),
    registration_date TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP NOT NULL,
    pvz_id UUID NOT NULL,
    product_ids JSONB DEFAULT '[]',
    status TEXT NOT NULL CHECK (status IN ('in_progress', 'close')),
    FOREIGN KEY (pvz_id) REFERENCES pvz(id)
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('электроника', 'одежда', 'обувь'))
    );

CREATE UNIQUE INDEX IF NOT EXISTS idx_receptions_pvz_status
    ON receptions(pvz_id, status) WHERE status = 'in_progress';

CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products(reception_id);   //можно еще индексы вынести