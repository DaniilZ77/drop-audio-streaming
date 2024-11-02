create table if not exists "beats" (
    "external_id" integer primary key,
    "beatmaker_id" integer not null,
    "path" varchar(64) not null,
    "artist" varchar(64) not null,
    "genre" varchar(64) not null,
    "is_deleted" boolean not null default false,
    "created_at" timestamp not null default current_timestamp,
    "updated_at" timestamp not null default current_timestamp
);

create index on "beats" ("artist");
create index on "beats" ("genre");