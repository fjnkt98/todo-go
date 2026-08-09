-- migrate:up
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME
);

-- migrate:down
DROP TABLE users;

