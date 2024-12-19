create table if not exists "beats" (
    "id" integer primary key,
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

create table if not exists "genres" (
    "id" serial primary key,
    "name" varchar(64) not null
);

create table if not exists "beats_genres" (
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "genre_id" integer not null references "genres" ("id")
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
create type "scale" as enum ('major', 'minor');

create table if not exists "beats_notes" (
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "note_id" integer not null references "notes" ("id"),
    "scale" scale not null
);

create table if not exists "beats_moods" (
    "id" serial primary key,
    "beat_id" integer not null references "beats" ("id"),
    "mood_id" integer not null references "moods" ("id")
);

insert into "beats" ("id", "beatmaker_id", "file_path", "image_path", "name", "description", "is_file_downloaded", "is_image_downloaded", "is_deleted", "created_at", "updated_at", "bpm")
values
    (1, 101, '/path/to/beat1', '', 'hip-hop vibes', 'a smooth hip-hop beat with relaxing vibes.', false, true, false, current_timestamp, current_timestamp, 100),
    (2, 102, '/path/to/beat2', '', 'lo-fi chill', 'a mellow lo-fi beat perfect for studying or relaxing.', true, true, false, current_timestamp, current_timestamp, 200),
    (3, 103, '/path/to/beat3', '', 'trap essentials', 'hard-hitting trap beat with a modern feel.', true, true, false, current_timestamp, current_timestamp, 300),
    (5, 104, '/path/to/beat4', '', 'synthwave journey', 'a retro synthwave beat with 80s vibes.', false, true, false, current_timestamp, current_timestamp, 400),
    (6, 104, '/path/to/beat5', '', 'jazzy night', 'a jazzy beat perfect for late night vibes.', false, true, false, current_timestamp, current_timestamp, 500),
    (7, 104, '/path/to/beat6', '', 'rock energy', 'high-energy rock beat for dynamic projects.', true, true, false, current_timestamp, current_timestamp, 600);


insert into "beats_genres" ("beat_id", "genre_id")
values
    (1, 1),
    (2, 2),
    (3, 3),
    (1, 4),
    (2, 5);

insert into "beats_tags" ("beat_id", "tag_id")
values
    (1, 1),
    (2, 2),
    (3, 3),
    (1, 4),
    (2, 5);

insert into "beats_moods" ("beat_id", "mood_id")
values
    (1, 1),
    (2, 2),
    (3, 3),
    (1, 4),
    (2, 5);

insert into "beats_notes" ("beat_id", "note_id", "scale")
values
    (1, 1, 'major'),
    (2, 2, 'minor'),
    (3, 3, 'major'),
    (1, 4, 'minor'),
    (2, 5, 'major');