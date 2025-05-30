-- name: CreatePoll :one
insert into poll(question, guild_id, author_id, duration, is_multi)
values ($1, $2, $3, $4, $5)
returning id;

-- name: CreatePollOption :exec
insert into poll_option(answer, emoji, poll_id)
VALUES ($1, $2, $3);

-- name: FindPollByID :one
select p.id,
       p.question,
       p.guild_id,
       p.author_id,
       p.is_multi,
       p.duration,
       p.created_at,
       array_agg(concat(po.emoji, '  ', po.answer)) :: text[] as options
from poll p
         left join poll_option po on p.id = po.poll_id
where p.id = $1
  and p.guild_id = $2
group by p.id;

-- name: FindPollByQuestion :many
select p.id,
       p.question,
       p.guild_id,
       p.author_id,
       p.is_multi,
       p.duration,
       p.created_at,
       array_agg(concat(po.emoji, '  ', po.answer)) :: text[] as options
from poll p
         left join poll_option po on p.id = po.poll_id
where p.question ilike concat('%', $1 :: text, '%')
  and p.guild_id = $2
group by p.id
offset $3 limit 10;

-- name: FindPollByIdAndQuestion :one
select p.id,
       p.question,
       p.guild_id,
       p.author_id,
       p.is_multi,
       p.duration,
       p.created_at,
       array_agg(concat(po.emoji, '  ', po.answer)) :: text[] as options
from poll p
         left join poll_option po on p.id = po.poll_id
where p.question ilike concat('%', $1 :: text, '%')
  and p.id = $2
  and p.guild_id = $3
group by p.id;

-- name: DeletePoll :exec
delete
from poll
where id = $1;

-- name: DeletePollOptions :exec
delete
from poll_option
where poll_id = $1;

-- name: ExistPoll :one
select true
from poll
where question = $1
  and guild_id = $2;