-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
  id integer PRIMARY KEY AUTOINCREMENT,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  name text,
  slug text,
  parent_id integer,
  position integer DEFAULT 0,
  is_locked integer DEFAULT 0,
  FOREIGN KEY(parent_id) REFERENCES categories(id) ON DELETE CASCADE
);
CREATE INDEX idx_categories_deleted_at ON categories(deleted_at);
CREATE INDEX idx_categories_parent_id ON categories(parent_id);

INSERT INTO categories (name, slug, position, is_locked, created_at, updated_at) VALUES ('Home', 'home', 0, 1, datetime('now'), datetime('now'));
INSERT INTO categories (name, slug, position, is_locked, created_at, updated_at) VALUES ('Contact', 'contact', 1, 1, datetime('now'), datetime('now'));

CREATE TABLE product_categories (
  product_id integer,
  category_id integer,
  PRIMARY KEY (product_id, category_id),
  FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE CASCADE,
  FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE product_categories;
DROP TABLE categories;
-- +goose StatementEnd
