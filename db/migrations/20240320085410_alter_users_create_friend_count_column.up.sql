ALTER TABLE users
ADD COLUMN IF NOT EXISTS friend_count INTEGER NOT NULL DEFAULT 0;
