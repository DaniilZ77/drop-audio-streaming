-- name: SaveBeat :exec
insert into beats ("id", "beatmaker_id", "bpm", "description", "name", "file_path", "image_path")
values ($1, $2, $3, $4, $5, $6, $7);

-- name: SaveGenres :copyfrom
insert into beats_genres ("beat_id", "genre_id")
values ($1, $2);

-- name: SaveTags :copyfrom
insert into beats_tags ("beat_id", "tag_id")
values ($1, $2);

-- name: SaveMoods :copyfrom
insert into beats_moods ("beat_id", "mood_id")
values ($1, $2);

-- name: SaveNote :exec
insert into beats_notes ("beat_id", "note_id", "scale")
values ($1, $2, $3);

-- name: GetBeatByID :one
select * from beats where id = $1;

-- name: GetBeatGenreParams :many
select * from genres;

-- name: GetBeatTagParams :many
select * from tags;

-- name: GetBeatMoodParams :many
select * from moods;

-- name: GetBeatNoteParams :many
select * from notes;

-- name: UpdateBeat :one
update beats
set "name" = coalesce(sqlc.narg('name'), "name"),
    "bpm" = coalesce(sqlc.narg('bpm'), "bpm"),
    "description" = coalesce(sqlc.narg('description'), "description"),
    "is_image_downloaded" = coalesce(sqlc.narg('is_image_downloaded'), "is_image_downloaded"),
    "is_file_downloaded" = coalesce(sqlc.narg('is_file_downloaded'), "is_file_downloaded")
where "id" = sqlc.arg('id') and "is_deleted" = false
returning *;

-- name: DeleteBeatGenres :exec
delete from beats_genres where beat_id = $1;

-- name: DeleteBeatTags :exec
delete from beats_tags where beat_id = $1;

-- name: DeleteBeatMoods :exec
delete from beats_moods where beat_id = $1;

-- name: DeleteBeatNotes :exec
delete from beats_notes where beat_id = $1;

-- name: DeleteBeat :exec
update beats set is_deleted = true where id = $1;