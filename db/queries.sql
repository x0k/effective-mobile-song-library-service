-- name: InsertSongAndReturnId :one
INSERT INTO
  song (title, artist, release_date, lyrics, link)
VALUES
  ($1, $2, $3, $4, $5)
RETURNING id;
