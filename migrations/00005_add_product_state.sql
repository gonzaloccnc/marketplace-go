-- +goose Up
-- +goose StatementBegin
ALTER TABLE products
    ADD COLUMN state CHAR(1) NOT NULL DEFAULT 'A' CHECK (state IN ('A', 'I'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS state;
-- +goose StatementEnd
