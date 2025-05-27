CREATE TABLE IF NOT EXISTS mentors (
    id              SERIAL PRIMARY KEY,
    first_name      VARCHAR(60)  NOT NULL,
    last_name       VARCHAR(60)  NOT NULL,
    about           VARCHAR(255) NOT NULL,
    track           VARCHAR(100) NOT NULL,
    education       VARCHAR(100) NOT NULL,
    experience_y    INT          NOT NULL,
    age             INT          NOT NULL,
    gender          VARCHAR(10)  NOT NULL,
    avatar_url      VARCHAR(255) NOT NULL,
    contact_url     VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT now()
    );
