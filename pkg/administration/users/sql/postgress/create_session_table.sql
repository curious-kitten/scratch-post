CREATE TABLE IF NOT EXISTS sessions (
    id serial primary key,
    username text not null,
    sessionid text not null unique,
    expirationTime timestamp not null
);