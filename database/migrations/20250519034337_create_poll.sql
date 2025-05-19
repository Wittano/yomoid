-- +goose Up
-- +goose StatementBegin
create table poll
(
    id         bigint primary key generated always as identity,
    question   varchar(300) not null check ( trim(question) <> '' ),
    guild_id   varchar      not null check ( trim(guild_id) <> '' ),
    author_id  varchar      not null check ( trim(author_id) <> ''),
    is_multi   bool         not null                                                   default false,
    duration   int2         not null check ( duration in (1, 4, 8, 24, 72, 168, 336) ) default 24,
    created_at timestamptz  not null                                                   default now()
);

create unique index poll_idx on poll (question, guild_id);

create table poll_option
(
    id         bigint primary key generated always as identity,
    answer     varchar(55) not null,
    emoji      varchar,
    created_at timestamptz not null default now(),
    poll_id    bigint      not null references poll
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists poll_option;
drop table if exists poll;
drop index if exists poll_idx;
-- +goose StatementEnd
