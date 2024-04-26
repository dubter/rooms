CREATE TABLE IF NOT EXISTS users(
   id serial PRIMARY KEY,
   nickname VARCHAR (50) UNIQUE NOT NULL,
   password_hash VARCHAR (100) NOT NULL,
   refresh_token VARCHAR (100),
   expires_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rooms(
  id serial PRIMARY KEY,
  name VARCHAR (50) NOT NULL,
  time_created TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages(
   id serial PRIMARY KEY,
   user_id INTEGER NOT NULL REFERENCES users(id),
   room_id INTEGER  NOT NULL REFERENCES rooms(id),
   content VARCHAR (300) NOT NULL,
   time_created TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_time_created ON messages (time_created);

CREATE INDEX IF NOT EXISTS idx_refresh_token ON users (refresh_token);

CREATE INDEX IF NOT EXISTS idx_nickname ON users (nickname);
