-- auth schema
CREATE SCHEMA IF NOT EXISTS storage;
CREATE SCHEMA IF NOT EXISTS auth;

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;
CREATE EXTENSION IF NOT EXISTS citext WITH SCHEMA public;

CREATE TABLE IF NOT EXISTS auth.users (
  id uuid DEFAULT public.gen_random_uuid () NOT NULL PRIMARY KEY
);
