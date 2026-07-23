-- +goose Up
-- +goose StatementBegin
CREATE TYPE user_created_via AS ENUM ('self_register', 'admin');
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN created_via user_created_via NOT NULL DEFAULT 'self_register',
    ADD COLUMN created_by  UUID NULL REFERENCES users (id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS created_via;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TYPE IF EXISTS user_created_via;
-- +goose StatementEnd
