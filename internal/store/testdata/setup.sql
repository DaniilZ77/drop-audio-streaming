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
    "updated_at" timestamp not null default current_timestamp
);
create table if not exists "beats_genres" (
    "id" serial primary key,
    "beat_id" integer not null references beats(id),
    "genre" varchar(64) not null
);

insert into "beats" ("id", "beatmaker_id", "file_path", "image_path", "name", "description", "is_file_downloaded", "is_image_downloaded", "is_deleted", "created_at", "updated_at")
values
    (1, 101, '/path/to/beat1', '', 'hip-hop vibes', 'a smooth hip-hop beat with relaxing vibes.', false, true, false, current_timestamp, current_timestamp),
    (2, 102, '/path/to/beat2', '', 'lo-fi chill', 'a mellow lo-fi beat perfect for studying or relaxing.', true, true, false, current_timestamp, current_timestamp),
    (3, 103, '/path/to/beat3', '', 'trap essentials', 'hard-hitting trap beat with a modern feel.', true, true, true, current_timestamp, current_timestamp),
    (5, 104, '/path/to/beat4', '', 'synthwave journey', 'a retro synthwave beat with 80s vibes.', false, true, false, current_timestamp, current_timestamp),
    (6, 104, '/path/to/beat5', '', 'jazzy night', 'a jazzy beat perfect for late night vibes.', false, true, false, current_timestamp, current_timestamp),
    (7, 104, '/path/to/beat6', '', 'rock energy', 'high-energy rock beat for dynamic projects.', true, true, false, current_timestamp, current_timestamp);


insert into "beats_genres" ("beat_id", "genre")
values
    (1, 'hip-hop'),
    (2, 'lo-fi'),
    (3, 'trap'),
    (1, 'chill'),
    (2, 'ambient');
