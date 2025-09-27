CREATE TABLE users
(
    id                BIGINT PRIMARY KEY,
    id_user           BIGINT unique,
    nickname_cashed   TEXT,
    name_cashed       TEXT,
    photo_uuid_cashed TEXT,
    is_dummy bool default false
);

CREATE TABLE category
(
    id          SERIAL PRIMARY KEY,
    name        TEXT,
    image_id    TEXT
);

CREATE TABLE event
(
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT,
    description TEXT,
    category_id SERIAL,
    image_id    TEXT,
    status      TEXT CHECK (status IN ('active', 'archive')),
    FOREIGN KEY (category_id) REFERENCES category (id)
);

CREATE TABLE user_event
(
    id_user       BIGINT NOT NULL,
    id_event      BIGINT NOT NULL,
    PRIMARY KEY (id_user, id_event),
    FOREIGN KEY (id_user) REFERENCES users (id_user) ON DELETE CASCADE,
    FOREIGN KEY (id_event) REFERENCES event (id) ON DELETE CASCADE
);

CREATE TABLE activity
(
    id          SERIAL PRIMARY KEY,
    id_event    INTEGER NOT NULL,
    id_user     INTEGER NOT NULL,
    description TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (id_event) REFERENCES event (id) ON DELETE CASCADE,
    FOREIGN KEY (id_user) REFERENCES users (id_user) ON DELETE CASCADE
);

CREATE TABLE task
(
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER,
    event_id    INTEGER NOT NULL,
    title       TEXT,
    description TEXT,
    priority    INTEGER,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id_user) ON DELETE SET NULL,
    FOREIGN KEY (event_id) REFERENCES event (id) ON DELETE CASCADE
);
