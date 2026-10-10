-- reverse: create index "idx_products_deleted_at" to table: "products"
DROP INDEX "idx_products_deleted_at";
-- reverse: create index "idx_products_code" to table: "products"
DROP INDEX "idx_products_code";
-- reverse: create "products" table
DROP TABLE "products";
-- reverse: create index "idx_jobs_type" to table: "jobs"
DROP INDEX "idx_jobs_type";
-- reverse: create index "idx_jobs_status" to table: "jobs"
DROP INDEX "idx_jobs_status";
-- reverse: create index "idx_jobs_requested_by" to table: "jobs"
DROP INDEX "idx_jobs_requested_by";
-- reverse: create index "idx_jobs_deleted_at" to table: "jobs"
DROP INDEX "idx_jobs_deleted_at";
-- reverse: create "jobs" table
DROP TABLE "jobs";
