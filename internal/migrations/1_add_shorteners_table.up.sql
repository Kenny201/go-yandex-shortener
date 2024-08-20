CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS shorteners
(
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    short_key    VARCHAR NOT NULL,
    original_url VARCHAR NOT NULL UNIQUE,
    user_id      UUID
);