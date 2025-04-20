-- +goose Up
ALTER TABLE users
    ADD hashed_password TEXT DEFAULT NULL;

-- +goose Down
ALTER TABLE users
    DROP COLUMN hashed_password;
