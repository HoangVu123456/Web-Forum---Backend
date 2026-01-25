-- Initial schema for WebForum
-- PostgreSQL dialect

CREATE TABLE IF NOT EXISTS users (
    user_id            BIGSERIAL PRIMARY KEY,
    username           VARCHAR(150) NOT NULL UNIQUE,
    email              VARCHAR(255) NOT NULL UNIQUE,
    password           VARCHAR(255) NOT NULL,
    profile_picture    VARCHAR(255),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tokens (
    token_id   BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token      VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS categories (
    category_id BIGSERIAL PRIMARY KEY,
    category    VARCHAR(150) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS posts (
    post_id     BIGSERIAL PRIMARY KEY,
    owner_id    BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    category_id BIGINT NOT NULL REFERENCES categories(category_id) ON DELETE CASCADE,
    headline    VARCHAR(255) NOT NULL,
    text        TEXT,
    image       VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status      BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS comments (
    comment_id         BIGSERIAL PRIMARY KEY,
    post_id            BIGINT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
    owner_id           BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    parent_comment_id  BIGINT REFERENCES comments(comment_id) ON DELETE SET NULL,
    text               TEXT NOT NULL,
    image              VARCHAR(255),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status             BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS reaction_types (
    reaction_type_id BIGSERIAL PRIMARY KEY,
    name             VARCHAR(100) NOT NULL UNIQUE,
    image            VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS reactions (
    reaction_id      BIGSERIAL PRIMARY KEY,
    post_id          BIGINT NOT NULL REFERENCES posts(post_id) ON DELETE CASCADE,
    owner_id         BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    reaction_type_id BIGINT NOT NULL REFERENCES reaction_types(reaction_type_id) ON DELETE RESTRICT,
    CONSTRAINT reactions_unique_owner_post UNIQUE (post_id, owner_id)
);

CREATE TABLE IF NOT EXISTS comment_reactions (
    comment_reaction_id BIGSERIAL PRIMARY KEY,
    comment_id          BIGINT NOT NULL REFERENCES comments(comment_id) ON DELETE CASCADE,
    owner_id            BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    reaction_type_id    BIGINT NOT NULL REFERENCES reaction_types(reaction_type_id) ON DELETE RESTRICT,
    CONSTRAINT comment_reactions_unique_owner_comment UNIQUE (comment_id, owner_id)
);

CREATE TABLE IF NOT EXISTS memberships (
    membership_id BIGSERIAL PRIMARY KEY,
    category_id   BIGINT NOT NULL REFERENCES categories(category_id) ON DELETE CASCADE,
    user_id       BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    joined_date   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT memberships_unique_user_category UNIQUE (category_id, user_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    notification_id   BIGSERIAL PRIMARY KEY,
    owner_id          BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    actor_id          BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    component_type    VARCHAR(50) NOT NULL,
    component_id      BIGINT NOT NULL,
    notification_type VARCHAR(50) NOT NULL,
    status            BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Helpful indexes for lookups
CREATE INDEX IF NOT EXISTS idx_tokens_user ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_owner ON posts(owner_id);
CREATE INDEX IF NOT EXISTS idx_posts_category ON posts(category_id);
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_owner ON comments(owner_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_comment_id);
CREATE INDEX IF NOT EXISTS idx_reactions_post_owner ON reactions(post_id, owner_id);
CREATE INDEX IF NOT EXISTS idx_comment_reactions_comment_owner ON comment_reactions(comment_id, owner_id);
CREATE INDEX IF NOT EXISTS idx_memberships_category_user ON memberships(category_id, user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_owner ON notifications(owner_id);
