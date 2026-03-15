-- +goose Up
-- +goose StatementBegin
CREATE TABLE settings (
  id integer PRIMARY KEY AUTOINCREMENT,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  key text UNIQUE,
  value text
);
CREATE INDEX idx_settings_deleted_at ON settings(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE settings;
-- +goose StatementEnd
