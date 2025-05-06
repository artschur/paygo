CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  name TEXT NOT NULL,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Then create wallets table (which references users)
CREATE TABLE wallets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
  balance BIGINT NOT NULL DEFAULT 0, -- stored in cents
  currency TEXT NOT NULL DEFAULT 'USD',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_wallets_user_id ON wallets (user_id);

-- Create accountants see
CREATE TABLE transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  from_wallet_id UUID REFERENCES wallets (id),
  to_wallet_id UUID REFERENCES wallets (id),
  amount BIGINT NOT NULL,
  status TEXT NOT NULL CHECK (
    status IN ('pending', 'completed', 'failed', 'refunded')
  ),
  type TEXT NOT NULL CHECK (type IN ('payment', 'refund', 'adjustment')),
  reference_id UUID, -- optional: could link to payments or refunds
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_transactions_from_wallet_id ON transactions (from_wallet_id);

CREATE INDEX idx_transactions_to_wallet_id ON transactions (to_wallet_id);

-- What the users see
CREATE TABLE payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
  sender_id UUID NOT NULL REFERENCES users (id),
  receiver_id UUID NOT NULL REFERENCES users (id),
  amount BIGINT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('initiated', 'completed', 'failed')),
  transaction_id UUID REFERENCES transactions (id),
  note TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
