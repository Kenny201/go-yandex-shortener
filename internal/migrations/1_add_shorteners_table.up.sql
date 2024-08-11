CREATE TABLE IF NOT EXISTS shorteners
(
    id           VARCHAR PRIMARY KEY,
    short_url    VARCHAR NOT NULL,
    original_url VARCHAR NOT NULL
);