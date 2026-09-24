-- Drop indexes
DROP INDEX IF EXISTS idx_support_messages_timestamp;
DROP INDEX IF EXISTS idx_support_messages_user;
DROP INDEX IF EXISTS idx_groups_creator;
DROP INDEX IF EXISTS idx_group_members_username;
DROP INDEX IF EXISTS idx_group_messages_group_timestamp;
DROP INDEX IF EXISTS idx_messages_receiver_sender_timestamp;
DROP INDEX IF EXISTS idx_messages_sender_receiver_timestamp;

-- Drop tables
DROP TABLE IF EXISTS support_messages;
DROP TABLE IF EXISTS group_messages;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS registration_keys;
DROP TABLE IF EXISTS users;
