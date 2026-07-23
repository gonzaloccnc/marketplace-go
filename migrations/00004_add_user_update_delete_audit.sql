-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN updated_by UUID NULL REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN deleted_by UUID NULL REFERENCES users (id) ON DELETE SET NULL;
-- +goose StatementEnd

-- Email uniqueness must apply only to live users so a soft-deleted user's email
-- can be reused. Swap the plain unique constraint for a partial unique index.
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX users_email_unique ON users (email) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_email_unique;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS deleted_by,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_by;
-- +goose StatementEnd
