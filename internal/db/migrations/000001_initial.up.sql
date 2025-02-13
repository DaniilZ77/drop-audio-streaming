create extension if not exists "uuid-ossp";

create table if not exists "beats" (
    "id" serial primary key,
    "beatmaker_id" integer not null,
    "file_path" varchar(64) not null,
    "image_path" varchar(64) not null,
    "name" varchar(128) not null,
    "description" text not null,
    "is_file_downloaded" boolean not null default false,
    "is_image_downloaded" boolean not null default false,
    "is_deleted" boolean not null default false,
    "created_at" timestamp not null default current_timestamp,
    "updated_at" timestamp not null default current_timestamp,
    "bpm" integer not null
);

create index on "beats" ("file_path");
create index on "beats" ("image_path");
create index on "beats" ("beatmaker_id");

create table if not exists "beats_genres" (
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "genre_id" integer not null references "genres" ("id")
);

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
    else
        update "beats"
        set "is_image_downloaded" = true
        where "image_path" = s3_path;
    end if;

    return new;
end;
$$;

create trigger trg_mark_downloaded
before insert on "beats_events"
for each row
execute function mark_downloaded();

create table if not exists "genres" (
    "id" serial primary key,
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

create table if not exists "tags" (
    "id" serial primary key,
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
    "id" serial primary key,
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
    "id" serial primary key,
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
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "tag_id" integer not null references "tags" ("id")
);

create index on "beats_tags" ("tag_id");

create type "scale" as enum ('major', 'minor');

create table if not exists "beats_notes" (
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "note_id" integer not null references "notes" ("id"),
    "scale" scale not null
);

create index on "beats_notes" ("note_id");
create unique index on "beats_notes" ("beat_id", "note_id");

create table if not exists "beats_moods" (
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "mood_id" integer not null references "moods" ("id")
);

create index on "beats_moods" ("mood_id");