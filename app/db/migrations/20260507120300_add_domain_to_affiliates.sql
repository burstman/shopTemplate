-- +goose Up
-- +goose StatementBegin
ALTER TABLE affiliates ADD COLUMN domain TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE affiliates DROP COLUMN domain;
-- +goose StatementEnd
