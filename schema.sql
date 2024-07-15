-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";

-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';

-- Create "items" table
CREATE TABLE "public"."items" (
    "id" bigserial NOT NULL,
    "title" TEXT NOT NULL,
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);
