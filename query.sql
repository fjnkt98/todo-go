-- name: FetchItems :many
SELECT
    "id",
    "title",
    "updated_at"
FROM
    "items"
ORDER BY
    "id" ASC;

-- name: CreateItem :one
INSERT INTO
    "items" ("title")
VALUES
    ($1)
RETURNING
    *;

-- name: UpdateItem :one
UPDATE "items"
SET
    "title" = $2,
    "updated_at" = NOW()
WHERE
    "id" = $1
RETURNING
    *;
