CREATE TABLE IF NOT EXISTS users (
   id serial primary key,
   email text not null unique,
   username text not null unique,
   name text not null,
   password text not null
);