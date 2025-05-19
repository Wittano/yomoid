-- name: CreatePoll :one
insert into poll(question, guild_id, author_id, duration, is_multi)
values ($1, $2, $3, $4, $5)
returning id;

-- name: CreatePollOption :exec
insert into poll_option(answer, emoji, poll_id)
VALUES ($1, $2, $3);