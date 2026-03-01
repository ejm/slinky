create table "links" (
  "id" serial primary key,
  "short_url" varchar(255) not null UNIQUE,
  "long_url" VARCHAR(255) not null,
  "vanity" BOOLEAN not null,
  "created" TIMESTAMPTZ not null default NOW()
);
