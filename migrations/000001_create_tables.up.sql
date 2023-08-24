CREATE TABLE users (
    id bigint PRIMARY KEY NOT NULL ,
    isactive bool NOT NULL
);

CREATE TABLE lichess (
    id SERIAL PRIMARY KEY,
    lichess_id TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    rapid INT NOT NULL,
    created_at TIMESTAMP(0) NOT NULL DEFAULT NOW()
);