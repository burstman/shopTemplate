-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
  id integer PRIMARY KEY AUTOINCREMENT,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  first_name text,
  last_name text,
  email text,
  address text,
  city text,
  phone text,
  total real,
  status text DEFAULT 'pending'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd
