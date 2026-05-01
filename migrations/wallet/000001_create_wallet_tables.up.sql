CREATE TABLE wallets (
    id VARCHAR PRIMARY KEY,
    user_id VARCHAR NOT NULL UNIQUE,
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    id VARCHAR PRIMARY KEY,
    from_wallet_id VARCHAR NOT NULL REFERENCES wallets(id),
    to_wallet_id VARCHAR NOT NULL REFERENCES wallets(id),
    amount BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);