/*
 * Copyright 2018 The Nakama Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    PRIMARY KEY (id),

    id            UUID          NOT NULL,
    username      VARCHAR(128)  CONSTRAINT users_username_key UNIQUE NOT NULL,
    display_name  VARCHAR(255),
    avatar_url    VARCHAR(512),
    -- https://tools.ietf.org/html/bcp47
    lang_tag      VARCHAR(18)   DEFAULT 'en',
    location      VARCHAR(255) DEFAULT 'any' NOT NULL, -- e.g. "San Francisco, CA"
    timezone      VARCHAR(255), -- e.g. "Pacific Time (US & Canada)"
    metadata      JSONB         DEFAULT '{}' NOT NULL,
    wallet        JSONB         DEFAULT '{}' NOT NULL,
    email         VARCHAR(255)  UNIQUE,
    password      BYTEA         CHECK (length(password) < 32000),
    facebook_id   VARCHAR(128)  UNIQUE,
    google_id     VARCHAR(128)  UNIQUE,
    gamecenter_id VARCHAR(128)  UNIQUE,
    steam_id      VARCHAR(128)  UNIQUE,
    custom_id     VARCHAR(128)  UNIQUE,
    role          INT           DEFAULT 0,
    edge_count    INT           DEFAULT 0 CHECK (edge_count >= 0) NOT NULL,
    last_login    TIMESTAMPTZ   DEFAULT NULL,
    create_time   TIMESTAMPTZ   DEFAULT now() NOT NULL,
    update_time   TIMESTAMPTZ   DEFAULT now() NOT NULL,
    verify_time   TIMESTAMPTZ   DEFAULT CAST(0 AS TIMESTAMPTZ) NOT NULL,
    disable_time  TIMESTAMPTZ   DEFAULT CAST(0 AS TIMESTAMPTZ) NOT NULL
);

CREATE INDEX IF NOT EXISTS last_login_idx ON users (last_login);


CREATE TABLE IF NOT EXISTS tournament (
    PRIMARY KEY (id),

    id            UUID          NOT NULL,
    name          string DEFAULT '',
    image_data    string default null,
    location      string default 'any',
    description   string default '',
    lang          int default 0,
    create_date    TIMESTAMPTZ default now() not null,
    start_date    TIMESTAMPTZ default now() not null,
    participants  int default 0,
    status        int default 0,
    current_round  UUID DEFAULT NULL,
    current_round_num  int DEFAULT 1
);

CREATE INDEX IF NOT EXISTS location_idx ON tournament (location);
CREATE INDEX IF NOT EXISTS start_date_idx ON tournament (start_date);


CREATE TABLE IF NOT EXISTS tour (
    PRIMARY KEY (id),
    FOREIGN KEY (tournament_id) REFERENCES tournament (id) ON DELETE CASCADE,

    id  UUID,
    tournament_id  UUID,
    end_date      TIMESTAMPTZ DEFAULT now(),
    award         int default 0,
    top_perc      float default 10
);

CREATE INDEX IF NOT EXISTS end_date_inx ON tour (end_date);

CREATE TABLE IF NOT EXISTS tour_user (
    PRIMARY KEY (id),
    id            UUID          NOT NULL,

    FOREIGN KEY (tour_id) REFERENCES tour (id) ON DELETE CASCADE,
    FOREIGN KEY (tournament_id) REFERENCES tournament (id) ON DELETE CASCADE,

    tournament_id  UUID,
    points   int default 0,
    submitted bool DEFAULT false,
    tour_id  UUID DEFAULT NULL,
    user_id UUID NOT NULL,
    status int default 0,
    award_money int default -1
);


CREATE TABLE IF NOT EXISTS matches (
    PRIMARY KEY (id),

    id            UUID          NOT NULL,
    user1         UUID          NOT NULL,
    user2         UUID          NOT NULL,
    win_user      UUID          DEFAULT NULL,
    waitingUser   UUID,
    round         int           DEFAULT 0,
    status        string,
    invite_id     UUID          NOT NULL,
    last_turn     TIMESTAMPTZ   DEFAULT NULL,
    user1_score   INT[],
    user2_score   INT[],
    words_seed    INT NOT NULL
);

CREATE INDEX IF NOT EXISTS invite_id_inx ON matches (invite_id);
CREATE INDEX IF NOT EXISTS user1_inx ON matches (user1);
CREATE INDEX IF NOT EXISTS user2_inx ON matches (user2);

-- Setup System user.
INSERT INTO users (id, username)
    VALUES ('00000000-0000-0000-0000-000000000000', '')
    ON CONFLICT(id) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_device (
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,

    id      VARCHAR(128) NOT NULL,
    user_id UUID         NOT NULL
);

CREATE TABLE IF NOT EXISTS category (
    PRIMARY KEY (id),

    id           UUID        NOT NULL,
    catname      VARCHAR(128),
    words_eng    STRING[],
    words_spa    STRING[],
    words_ita    STRING[],
    words_por    STRING[],
    words_fre    STRING[],
    points_eng   INT[],
    points_spa   INT[],
    points_ita   INT[],
    points_por   INT[],
    points_fre   INT[],
    image_data   STRING,
    description  string default '',
    unlock_price SMALLINT
);

CREATE TABLE IF NOT EXISTS countries (
    PRIMARY KEY (country),

    country string NOT NULL
);

CREATE TABLE IF NOT EXISTS settings (
    PRIMARY KEY (id),

    id                  UUID        NOT NULL,
    swag_levels         STRING[],
    swag_levels_desc    STRING[],
    swag_levels_points  INT[],
    len_round           INT DEFAULT 90,
    time_penalty_sec    INT DEFAULT 1,
    coins_per_point     FLOAT,
    h2hwords            INT DEFAULT 30,
    word_showup_time    float default 2
);

insert into settings (id, coins_per_point, swag_levels, swag_levels_points) VALUES('00000000-0000-0000-0000-000000000000', 0.5, ARRAY['test1','test2','test3'], ARRAY[10, 20, 30]);

CREATE TABLE IF NOT EXISTS unlocked_categories (
    PRIMARY KEY (id) ,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE CASCADE,

    id                UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id           UUID        NOT NULL,
    category_id       UUID        NOT NULL
);

CREATE TABLE IF NOT EXISTS purchases (
    PRIMARY KEY (id) ,
    FOREIGN KEY (user_id) REFERENCES users (id),

    id                UUID        NOT NULL DEFAULT gen_random_uuid(),
    create_time TIMESTAMPTZ  DEFAULT now() NOT NULL,
    item string not null,
    amount int default 1 not null,
    receipt string default null,
    user_id     UUID         NOT NULL
);

CREATE TABLE IF NOT EXISTS user_edge (
    PRIMARY KEY (source_id, state, position),
    FOREIGN KEY (source_id)      REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (destination_id) REFERENCES users (id) ON DELETE CASCADE,

    source_id      UUID        NOT NULL,
    position       BIGINT      NOT NULL, -- Used for sort order on rows.
    update_time    TIMESTAMPTZ DEFAULT now() NOT NULL,
    destination_id UUID        NOT NULL,
    state          SMALLINT    DEFAULT 0 NOT NULL, -- friend(0), invite(1), invited(2), blocked(3), deleted(4), archived(5)

    UNIQUE (source_id, destination_id)
);

CREATE TABLE IF NOT EXISTS notification (
    -- FIXME: cockroach's analyser is not clever enough when create_time has DESC mode on the index.
    PRIMARY KEY (user_id, create_time, id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,

    id          UUID         CONSTRAINT notification_id_key UNIQUE NOT NULL,
    user_id     UUID         NOT NULL,
    subject     VARCHAR(255) NOT NULL,
    content     JSONB        DEFAULT '{}' NOT NULL,
    code        SMALLINT     NOT NULL, -- Negative values are system reserved.
    sender_id   UUID         NOT NULL,
    create_time TIMESTAMPTZ  DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS storage (
    PRIMARY KEY (collection, read, key, user_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,

    collection  VARCHAR(128) NOT NULL,
    key         VARCHAR(128) NOT NULL,
    user_id     UUID         NOT NULL,
    value       JSONB        DEFAULT '{}' NOT NULL,
    version     VARCHAR(32)  NOT NULL, -- md5 hash of value object.
    read        SMALLINT     DEFAULT 1 CHECK (read >= 0) NOT NULL,
    write       SMALLINT     DEFAULT 1 CHECK (write >= 0) NOT NULL,
    create_time TIMESTAMPTZ  DEFAULT now() NOT NULL,
    update_time TIMESTAMPTZ  DEFAULT now() NOT NULL,

    UNIQUE (collection, key, user_id)
);
CREATE INDEX IF NOT EXISTS collection_read_user_id_key_idx ON storage (collection, read, user_id, key);
CREATE INDEX IF NOT EXISTS value_ginidx ON storage USING GIN (value);

CREATE TABLE IF NOT EXISTS message (
  PRIMARY KEY (stream_mode, stream_subject, stream_descriptor, stream_label, create_time, id),
  FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE,

  id                UUID         UNIQUE NOT NULL,
  -- chat(0), chat_update(1), chat_remove(2), group_join(3), group_add(4), group_leave(5), group_kick(6), group_promoted(7)
  code              SMALLINT     DEFAULT 0 NOT NULL,
  sender_id         UUID         NOT NULL,
  username          VARCHAR(128) NOT NULL,
  stream_mode       SMALLINT     NOT NULL,
  stream_subject    UUID         NOT NULL,
  stream_descriptor UUID         NOT NULL,
  stream_label      VARCHAR(128) NOT NULL,
  content           JSONB        DEFAULT '{}' NOT NULL,
  create_time       TIMESTAMPTZ  DEFAULT now() NOT NULL,
  update_time       TIMESTAMPTZ  DEFAULT now() NOT NULL,

  UNIQUE (sender_id, id)
);

CREATE TABLE IF NOT EXISTS achs (
  PRIMARY KEY (id),

  id          UUID        UNIQUE NOT NULL DEFAULT gen_random_uuid(),
  name          VARCHAR(255) NOT NULL,
  event         VARCHAR(255) NOT NULL,
  condition     SMALLINT NOT NULL,
  reward        SMALLINT NOT NULL
);



CREATE TABLE IF NOT EXISTS leaderboard (
  PRIMARY KEY (id),

  id             VARCHAR(128) NOT NULL,
  authoritative  BOOLEAN      DEFAULT FALSE,
  sort_order     SMALLINT     DEFAULT 1 NOT NULL, -- asc(0), desc(1)
  operator       SMALLINT     DEFAULT 0 NOT NULL, -- best(0), set(1), increment(2), decrement(3)
  reset_schedule VARCHAR(64), -- e.g. cron format: "* * * * * * *"
  metadata       JSONB        DEFAULT '{}' NOT NULL,
  create_time    TIMESTAMPTZ  DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS leaderboard_record (
  PRIMARY KEY (leaderboard_id, expiry_time, score, subscore, owner_id),
  FOREIGN KEY (leaderboard_id) REFERENCES leaderboard (id) ON DELETE CASCADE,

  leaderboard_id VARCHAR(128)  NOT NULL,
  owner_id       UUID          NOT NULL,
  username       VARCHAR(128),
  score          BIGINT        DEFAULT 0 CHECK (score >= 0) NOT NULL,
  subscore       BIGINT        DEFAULT 0 CHECK (subscore >= 0) NOT NULL,
  num_score      INT           DEFAULT 1 CHECK (num_score >= 0) NOT NULL,
  metadata       JSONB         DEFAULT '{}' NOT NULL,
  create_time    TIMESTAMPTZ   DEFAULT now() NOT NULL,
  update_time    TIMESTAMPTZ   DEFAULT now() NOT NULL,
  expiry_time    TIMESTAMPTZ   DEFAULT CAST(0 AS TIMESTAMPTZ) NOT NULL,

  UNIQUE (owner_id, leaderboard_id, expiry_time)
);

CREATE TABLE IF NOT EXISTS passreq (
  PRIMARY KEY (id),
  FOREIGN KEY (user_id) REFERENCES users (id),

  id          UUID        UNIQUE NOT NULL DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL,
  code        VARCHAR(10)  UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS wallet_ledger (
  PRIMARY KEY (user_id, create_time, id),
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,

  id          UUID        UNIQUE NOT NULL,
  user_id     UUID        NOT NULL,
  changeset   JSONB       NOT NULL,
  metadata    JSONB       NOT NULL,
  create_time TIMESTAMPTZ DEFAULT now() NOT NULL,
  update_time TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS user_tombstone (
  PRIMARY KEY (create_time, user_id),

  user_id        UUID        UNIQUE NOT NULL,
  create_time    TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE IF NOT EXISTS groups (
  PRIMARY KEY (disable_time, lang_tag, edge_count, id),

  id           UUID          UNIQUE NOT NULL,
  creator_id   UUID          NOT NULL,
  name         VARCHAR(255)  CONSTRAINT groups_name_key UNIQUE NOT NULL,
  description  VARCHAR(255),
  avatar_url   VARCHAR(512),
  -- https://tools.ietf.org/html/bcp47
  lang_tag     VARCHAR(18)   DEFAULT 'en',
  metadata     JSONB         DEFAULT '{}' NOT NULL,
  state        SMALLINT      DEFAULT 0 CHECK (state >= 0) NOT NULL, -- open(0), closed(1)
  edge_count   INT           DEFAULT 0 CHECK (edge_count >= 1 AND edge_count <= max_count) NOT NULL,
  max_count    INT           DEFAULT 100 CHECK (max_count >= 1) NOT NULL,
  create_time  TIMESTAMPTZ   DEFAULT now() NOT NULL,
  update_time  TIMESTAMPTZ   DEFAULT now() NOT NULL,
  disable_time TIMESTAMPTZ   DEFAULT CAST(0 AS TIMESTAMPTZ) NOT NULL
);
CREATE INDEX IF NOT EXISTS edge_count_update_time_id_idx ON groups (disable_time, edge_count, update_time, id);
CREATE INDEX IF NOT EXISTS update_time_edge_count_id_idx ON groups (disable_time, update_time, edge_count, id);

CREATE TABLE IF NOT EXISTS group_edge (
  PRIMARY KEY (source_id, state, position),

  source_id      UUID        NOT NULL,
  position       BIGINT      NOT NULL, -- Used for sort order on rows.
  update_time    TIMESTAMPTZ DEFAULT now() NOT NULL,
  destination_id UUID        NOT NULL,
  state          SMALLINT    DEFAULT 0 NOT NULL, -- superadmin(0), admin(1), member(2), join_request(3), archived(4)

  UNIQUE (source_id, destination_id)
);

-- +migrate Down
DROP TABLE IF EXISTS group_edge;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS user_tombstone;
DROP TABLE IF EXISTS wallet_ledger;
DROP TABLE IF EXISTS leaderboard_record;
DROP TABLE IF EXISTS leaderboard;
DROP TABLE IF EXISTS message;
DROP TABLE IF EXISTS storage;
DROP TABLE IF EXISTS notification;
DROP TABLE IF EXISTS user_edge;
DROP TABLE IF EXISTS user_device;
DROP TABLE IF EXISTS users;
