CREATE TABLE IF NOT EXISTS user_friends (
  id SERIAL PRIMARY KEY,
  user_id_1 VARCHAR(48) NOT NULL,
  user_id_2 VARCHAR(48) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'confirmed',
  created_at TIMESTAMP(0) DEFAULT NOW(),
  updated_at TIMESTAMP(0) DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_friends_user_id_1 ON user_friends(user_id_1);
CREATE INDEX IF NOT EXISTS idx_user_friends_user_id_2 ON user_friends(user_id_2);
