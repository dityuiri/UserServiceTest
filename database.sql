CREATE TABLE IF NOT EXISTS user_master (
    id   UUID  PRIMARY KEY,
    phone_number VARCHAR(13) NOT NULL,
    name VARCHAR(60) NOT NULL,
    password_hash TEXT NOT NULL,

    CONSTRAINT phone_number_key UNIQUE(phone_number)
);

CREATE INDEX idx_user_phone_number ON user_master(phone_number);

CREATE TABLE IF NOT EXISTS user_login (
    user_id         UUID   PRIMARY KEY,
    successful_login INT   DEFAULT 0
);


