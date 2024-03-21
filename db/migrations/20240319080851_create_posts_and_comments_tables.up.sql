CREATE TABLE IF NOT EXISTS posts (
  id VARCHAR(48) PRIMARY KEY,
  user_id VARCHAR(48) NOT NULL,
  post_in_html TEXT NOT NULL,
  created_at TIMESTAMP(0) DEFAULT NOW(),
  updated_at TIMESTAMP(0) DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);

CREATE TABLE IF NOT EXISTS post_comments (
  id VARCHAR(48) PRIMARY KEY,
  user_id VARCHAR(48) NOT NULL,
  post_id VARCHAR(48) NOT NULL,
  comment TEXT NOT NULL,
  created_at TIMESTAMP(0) DEFAULT NOW(),
  updated_at TIMESTAMP(0) DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_post_comments_post_id ON post_comments(post_id);
CREATE INDEX IF NOT EXISTS idx_post_comments_user_id ON post_comments(user_id);

CREATE TABLE IF NOT EXISTS post_tags (
  id SERIAL PRIMARY KEY,
  post_id VARCHAR(48) NOT NULL,
  tag VARCHAR(32) NOT NULL,
  created_at TIMESTAMP(0) DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_post_tags_post_id ON post_tags(post_id);
CREATE INDEX IF NOT EXISTS idx_post_tags_tag ON post_tags(tag);
