-- create "products" table
CREATE TABLE "products" (
  "id" text NOT NULL DEFAULT gen_random_uuid(),
  "name" text NOT NULL,
  "code" text NOT NULL,
  "price" numeric NOT NULL DEFAULT 0,
  "image" text NULL,
  "description" text NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_user" text NULL,
  "last_modified_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_modified_user" text NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_products_code" to table: "products"
CREATE UNIQUE INDEX "idx_products_code" ON "products" ("code");
-- create index "idx_products_deleted_at" to table: "products"
CREATE INDEX "idx_products_deleted_at" ON "products" ("deleted_at");
