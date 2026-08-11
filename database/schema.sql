CREATE TABLE IF NOT EXISTS USERPOOL (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER DEFAULT (unixepoch()),
    user_id INTEGER NOT NULL UNIQUE,
    user_group TEXT DEFAULT 'user',
    display_language TEXT,
    query_language TEXT
); 