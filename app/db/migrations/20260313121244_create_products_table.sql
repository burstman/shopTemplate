-- +goose Up
-- +goose StatementBegin
CREATE TABLE products (
  id integer PRIMARY KEY AUTOINCREMENT,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  name text,
  price real,
  image text,
  category text,
  stock integer DEFAULT 0,
  promotion_price real DEFAULT 0,
  description text
);
CREATE INDEX idx_products_deleted_at ON products(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE products;
-- +goose StatementEnd
