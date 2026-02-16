-- +goose Up
ALTER TABLE user_profile_behavior ADD COLUMN std_dev_transaction_amount INTEGER DEFAULT 0;

-- +goose Down
ALTER TABLE user_profile_behavior DROP COLUMN std_dev_transaction_amount;
