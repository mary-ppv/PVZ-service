CREATE UNIQUE INDEX IF NOT EXISTS idx_receptions_pvz_status
    ON receptions(pvz_id, status) WHERE status = 'in_progress';

CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products(reception_id); 