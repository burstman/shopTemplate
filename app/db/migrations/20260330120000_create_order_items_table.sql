-- +goose Up
-- +goose StatementBegin
CREATE TABLE order_items (
  id integer PRIMARY KEY AUTOINCREMENT,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  order_id integer,
  product_id integer,
  product_name text,
  product_image text,
  quantity integer,
  price real,
  FOREIGN KEY(order_id) REFERENCES orders(id) ON DELETE CASCADE,
  FOREIGN KEY(product_id) REFERENCES products(id)
);
CREATE INDEX idx_order_items_deleted_at ON order_items(deleted_at);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE order_items;
-- +goose StatementEnd