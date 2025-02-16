create extension if not exists "uuid-ossp";

create table if not exists "beats" (
    "id" uuid primary key default uuid_generate_v4(),
    "beatmaker_id" uuid not null,
    "file_path" varchar(64) not null,
    "image_path" varchar(64) not null,
    "archive_path" varchar(64) not null,
    "name" varchar(128) not null,
    "description" text not null,
    "is_file_downloaded" boolean not null default false,
    "is_image_downloaded" boolean not null default false,
    "is_archive_downloaded" boolean not null default false,
    "range_start" bigint not null,
    "range_end" bigint not null,
    "is_deleted" boolean not null default false,
    "created_at" timestamp not null default current_timestamp,
    "updated_at" timestamp not null default current_timestamp,
    "bpm" integer not null
);

create index on "beats" ("file_path");
create index on "beats" ("image_path");
create index on "beats" ("archive_path");
create index on "beats" ("beatmaker_id");
create index on "beats" ("name");
create index on "beats" ("bpm");

create table if not exists "beats_events" (
    "event_time" timestamp not null,
    "event_data" jsonb not null
);

create or replace function mark_downloaded()
    returns trigger
    language plpgsql
as $$
declare
    s3_path text;
begin
    s3_path := NEW.event_data->'Records'->0->'s3'->'object'->>'key';

    if exists(select 1 from "beats" where "file_path" = s3_path) then
        update "beats"
        set "is_file_downloaded" = true
        where "file_path" = s3_path;
    elsif exists(select 1 from "beats" where "image_path" = s3_path) then
        update "beats"
        set "is_image_downloaded" = true
        where "image_path" = s3_path;
    else
        update "beats"
        set "is_archive_downloaded" = true
        where "archive_path" = s3_path;
    end if;

    return new;
end;
$$;

create trigger trg_mark_downloaded
before insert on "beats_events"
for each row
execute function mark_downloaded();

create table if not exists "genres" (
    "id" uuid primary key default uuid_generate_v4(),
    "name" varchar(64) not null
);

insert into
    "genres" ("name")
values
    ('Hip-hop'),
    ('Trap'),
    ('Rnb'),
    ('Pop'),
    ('Electronic'),
    ('House'),
    ('Lo-Fi'),
    ('Drill'),
    ('Techno'),
    ('UK Garage'),
    ('Drum and Bass'),
    ('Jungle'),
    ('Hyperpop');

create table if not exists "beats_genres" (
    "id" uuid primary key default uuid_generate_v4(),
    "beat_id" uuid not null references "beats" ("id"),
    "genre_id" uuid not null references "genres" ("id")
);

create index on "beats_genres" ("beat_id");

create table if not exists "tags" (
    "id" uuid primary key default uuid_generate_v4(),
    "name" varchar(64) not null
);

insert into
    "tags" ("name")
values
    ('808'),
    ('trap'),
    ('drake'),
    ('lil baby'),
    ('gunna'),
    ('type beat'),
    ('guitar'),
    ('juice wrld'),
    ('rap'),
    ('hip hop');

create table if not exists "notes" (
    "id" uuid primary key default uuid_generate_v4(),
    "name" varchar(64) not null
);

insert into
    "notes" ("name")
values
    ('C'),
    ('C#'),
    ('D'),
    ('D#'),
    ('E'),
    ('F'),
    ('F#'),
    ('G'),
    ('G#'),
    ('A'),
    ('A#'),
    ('B');

create table if not exists "moods" (
    "id" uuid primary key default uuid_generate_v4(),
    "name" varchar(64) not null
);

insert into
    "moods" ("name")
values
    ('качовый'),
    ('темный'),
    ('меланхоличный'),
    ('лиричный'),
    ('спокойный'),
    ('агрессивный'),
    ('грустный'),
    ('депрессивный'),
    ('энергичный');

create table if not exists "beats_tags" (
    "id" uuid primary key default uuid_generate_v4(),
    "beat_id" uuid not null references "beats" ("id"),
    "tag_id" uuid not null references "tags" ("id")
);

create index on "beats_tags" ("beat_id");

create type "note_scale" as enum ('major', 'minor');

create table if not exists "beats_notes" (
    "id" uuid primary key default uuid_generate_v4(),
    "beat_id" uuid not null references "beats" ("id"),
    "note_id" uuid not null references "notes" ("id"),
    "scale" note_scale not null
);

create index on "beats_notes" ("beat_id");
create unique index on "beats_notes" ("beat_id", "note_id");

create table if not exists "beats_moods" (
    "id" uuid primary key default uuid_generate_v4(),
    "beat_id" uuid not null references "beats" ("id"),
    "mood_id" uuid not null references "moods" ("id")
);

create index on "beats_moods" ("beat_id");

create table "beats_owners" (
  "beat_id" uuid primary key,
  "user_id" uuid not null
);

alter table "beats_owners" add foreign key ("beat_id") references "beats" ("id");