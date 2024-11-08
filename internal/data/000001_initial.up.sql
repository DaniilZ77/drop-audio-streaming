create table if not exists "beats" (
    "id" integer primary key,
    "beatmaker_id" integer not null,
    "path" varchar(64) not null,
    "name" varchar(128) not null,
    "description" text not null,
    "is_downloaded" boolean not null default false,
    "is_deleted" boolean not null default false,
    "created_at" timestamp not null default current_timestamp,
    "updated_at" timestamp not null default current_timestamp
);

create index on "beats" ("path");
create index on "beats" ("beatmaker_id");

create table if not exists "beats_genres" (
    "id" serial primary key,
    "beat_id" integer not null,
    "genre" varchar(64) not null
);

create index on "beats_genres" ("genre");

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

    update "beats"
    set "is_downloaded" = true
    where "path" = s3_path;

    return new;
end;
$$;

create trigger trg_mark_downloaded
before insert on "beats_events"
for each row
execute function mark_downloaded();