CREATE TABLE IF NOT EXISTS shorteners
(
    id           uuid DEFAULT gen_random_uuid()  PRIMARY KEY,
    short_url    VARCHAR NOT NULL,
    original_url VARCHAR NOT NULL
);