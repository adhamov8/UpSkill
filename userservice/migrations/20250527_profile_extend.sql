-- userservice/migrations/20250527_profile_extend.sql
ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS first_name VARCHAR(60),
    ADD COLUMN IF NOT EXISTS last_name  VARCHAR(60),
    ADD COLUMN IF NOT EXISTS email      VARCHAR(120);
