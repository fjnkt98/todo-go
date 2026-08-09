CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20260809033946'),
  ('20260809034157');
