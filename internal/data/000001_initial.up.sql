create table if not exists "beats" (
    "id" serial primary key,
    "user_id" integer not null,
    "path" text not null,
    "is_deleted" boolean not null default false,
    "created_at" timestamp not null default current_timestamp,
    "updated_at" timestamp not null default current_timestamp
);