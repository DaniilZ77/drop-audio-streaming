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

create table if not exists "beats_moods" (
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "mood_id" integer not null references "moods" ("id")
);

create index on "beats_moods" ("mood_id");

alter table "beats_genres"
add column "genre_id" integer not null;

alter table "beats_genres" add foreign key ("genre_id") references "genres" ("id");

create index on "beats_genres" ("genre_id");

alter table "beats_genres"
drop column "genre";

alter table "beats"
add column "bpm" integer not null;

alter table "beats_genres"
add constraint "beats_genres_beat_id_fkey"
foreign key ("beat_id") references "beats" ("id");