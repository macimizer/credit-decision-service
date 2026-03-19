CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    birth_date DATE NOT NULL,
    country TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS banks (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('PRIVATE', 'GOVERNMENT'))
);

CREATE TABLE IF NOT EXISTS credits (
    id UUID PRIMARY KEY,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    bank_id UUID NOT NULL REFERENCES banks(id) ON DELETE RESTRICT,
    min_payment NUMERIC(12,2) NOT NULL,
    max_payment NUMERIC(12,2) NOT NULL,
    term_months INT NOT NULL,
    credit_type TEXT NOT NULL CHECK (credit_type IN ('AUTO', 'MORTGAGE', 'COMMERCIAL')),
    created_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED'))
);

CREATE INDEX IF NOT EXISTS idx_clients_email ON clients(email);
CREATE INDEX IF NOT EXISTS idx_credits_client_id ON credits(client_id);
CREATE INDEX IF NOT EXISTS idx_credits_bank_id ON credits(bank_id);
CREATE INDEX IF NOT EXISTS idx_credits_status ON credits(status);
