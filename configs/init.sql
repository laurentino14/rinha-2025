CREATE UNLOGGED TABLE payments (
    correlation_id UUID PRIMARY KEY,
    amount        DECIMAL NOT NULL,
    processor SMALLINT NOT NULL CHECK (processor IN (0, 1)),
    requested_at   TIMESTAMP NOT NULL
);

CREATE INDEX idx_payments_processor_requested_at ON payments (requested_at);
