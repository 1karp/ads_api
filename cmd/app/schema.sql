CREATE TABLE IF NOT EXISTS users (
    userid INTEGER PRIMARY KEY,
    ads TEXT,
    username TEXT
);

CREATE TABLE IF NOT EXISTS ads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    username TEXT,
    photos TEXT NOT NULL,
    rooms TEXT,
    price INTEGER,
    type TEXT,
    area INTEGER,
    building TEXT,
    district TEXT,
    text TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_posted INTEGER DEFAULT 0,
    chat_message_id INTEGER,
    FOREIGN KEY(user_id) REFERENCES users(userid)
);
