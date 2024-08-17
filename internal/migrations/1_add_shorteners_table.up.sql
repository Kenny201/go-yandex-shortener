CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS shorteners
(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    short_url    VARCHAR NOT NULL,
    original_url VARCHAR NOT NULL UNIQUE
);