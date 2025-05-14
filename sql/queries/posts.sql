-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, title, url, feed_id, description, published_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.*, feeds.name AS feed_name, users.name AS user_name
FROM posts
JOIN feeds ON posts.feed_id = feeds.id
JOIN users ON feeds.user_id = users.id
WHERE feeds.user_id = $1
ORDER BY posts.created_at DESC
LIMIT $2;