-- Users table
CREATE TABLE users (
    username VARCHAR(32) PRIMARY KEY,
    account_id VARCHAR(8) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    middle_name VARCHAR(100),
    display_name VARCHAR(100) NOT NULL,
    is_admin BOOLEAN DEFAULT FALSE,
    is_blocked BOOLEAN DEFAULT FALSE,
    bio VARCHAR(30) DEFAULT '',
    avatar_key VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login TIMESTAMPTZ
);

-- Registration keys table
CREATE TABLE registration_keys (
    key VARCHAR(16) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    used_by VARCHAR(32) REFERENCES users(username) ON DELETE SET NULL,
    used_at TIMESTAMPTZ
);

-- Messages table (P2P)
CREATE TABLE messages (
    id VARCHAR(8) PRIMARY KEY,
    sender VARCHAR(32) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    receiver VARCHAR(32) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    content TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type VARCHAR(20) NOT NULL DEFAULT 'text' CHECK (type IN ('text', 'image', 'file', 'saved')),
    edited BOOLEAN DEFAULT FALSE,
    deleted BOOLEAN DEFAULT FALSE,
    read BOOLEAN DEFAULT FALSE,
    reply_to VARCHAR(8) REFERENCES messages(id) ON DELETE SET NULL,
    reactions JSONB NOT NULL DEFAULT '{}'::jsonb,
    CHECK (
        (type != 'saved') OR (sender = receiver)
    )
);

-- Groups table
CREATE TABLE groups (
    id VARCHAR(8) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    creator VARCHAR(32) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settings JSONB NOT NULL DEFAULT '{"allow_reactions": true}'::jsonb
);

-- Group members table
CREATE TABLE group_members (
    group_id VARCHAR(8) REFERENCES groups(id) ON DELETE CASCADE,
    username VARCHAR(32) REFERENCES users(username) ON DELETE CASCADE,
    PRIMARY KEY (group_id, username)
);

-- Group messages table
CREATE TABLE group_messages (
    id VARCHAR(8) PRIMARY KEY,
    group_id VARCHAR(8) NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    sender VARCHAR(32) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    content TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type VARCHAR(20) NOT NULL DEFAULT 'group' CHECK (type IN ('group', 'image', 'file')),
    edited BOOLEAN DEFAULT FALSE,
    deleted BOOLEAN DEFAULT FALSE,
    reply_to VARCHAR(8) REFERENCES group_messages(id) ON DELETE SET NULL,
    reactions JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- Support messages table
CREATE TABLE support_messages (
    id VARCHAR(8) PRIMARY KEY,
    user VARCHAR(32) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    sender VARCHAR(32) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    content TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (type IN ('user', 'admin')),
    answered BOOLEAN DEFAULT FALSE
);

-- Indexes for pagination
CREATE INDEX idx_messages_sender_receiver_timestamp ON messages(sender, receiver, timestamp DESC);
CREATE INDEX idx_messages_receiver_sender_timestamp ON messages(receiver, sender, timestamp DESC);
CREATE INDEX idx_group_messages_group_timestamp ON group_messages(group_id, timestamp DESC);
CREATE INDEX idx_group_members_username ON group_members(username);
CREATE INDEX idx_groups_creator ON groups(creator);
CREATE INDEX idx_support_messages_user ON support_messages(user);
CREATE INDEX idx_support_messages_timestamp ON support_messages(timestamp DESC);
