-- create "jobs" table
CREATE TABLE "jobs" (
  "id" text NOT NULL DEFAULT gen_random_uuid(),
  "type" text NOT NULL,
  "status" text NOT NULL DEFAULT 'PENDING',
  "payload" jsonb NULL,
  "result" jsonb NULL,
  "error_message" text NULL,
  "requested_by" text NULL,
  "completed_at" timestamptz NULL,
  "retry_count" bigint NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "created_user" text NULL,
  "last_modified_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "last_modified_user" text NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "idx_jobs_deleted_at" to table: "jobs"
CREATE INDEX "idx_jobs_deleted_at" ON "jobs" ("deleted_at");
-- create index "idx_jobs_requested_by" to table: "jobs"
CREATE INDEX "idx_jobs_requested_by" ON "jobs" ("requested_by");
-- create index "idx_jobs_status" to table: "jobs"
CREATE INDEX "idx_jobs_status" ON "jobs" ("status");
-- create index "idx_jobs_type" to table: "jobs"
CREATE INDEX "idx_jobs_type" ON "jobs" ("type");
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
