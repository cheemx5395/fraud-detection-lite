-- +goose Up

CREATE TYPE mode AS ENUM (
  'UPI',
  'CARD',
  'NETBANKING'
);

CREATE TYPE trigger_factors AS ENUM (
  'AMOUNT_DEVIATION',
  'FREQUENCY_SPIKE',
  'NEW_MODE',
  'TIME_ANOMALY'
);

CREATE TYPE transaction_decision AS ENUM (
  'ALLOW',
  'FLAG',
  'BLOCK',
  'MFA_REQUIRED'
);

CREATE TYPE transaction_type AS ENUM (
  'CREDIT',
  'DEBIT'
);

CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  mobile_number VARCHAR(20) UNIQUE NOT NULL,
  hashed_pass VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_profile_behavior (
  user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  average_transaction_amount INTEGER,
  max_transaction_amount_seen INTEGER,
  average_transactions_per_day INTEGER,
  registered_payment_modes mode[] NOT NULL DEFAULT '{}',
  usual_transaction_start_hour TIMESTAMP,
  usual_transaction_end_hour TIMESTAMP,
  total_transactions INTEGER NOT NULL DEFAULT 0,
  allowed_transactions INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE transactions (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount INTEGER NOT NULL,
  type transaction_type NOT NULL,
  mode mode NOT NULL,
  risk_score INTEGER NOT NULL,
  triggered_factors trigger_factors[] NOT NULL DEFAULT '{}',
  decision transaction_decision NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down

DROP TABLE IF EXISTS transactions;

DROP TABLE IF EXISTS user_profile_behavior;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS transaction_decision;

DROP TYPE IF EXISTS trigger_factors;

DROP TYPE IF EXISTS mode;

DROP TYPE IF EXISTS transaction_type;