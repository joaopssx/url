CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS urls (
    id TEXT PRIMARY KEY,
    short_code TEXT UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    user_id TEXT,
    created_at DATETIME NOT NULL,
    expires_at DATETIME,
    access_count INTEGER DEFAULT 0,
    deleted_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS url_accesses (
    id TEXT PRIMARY KEY,
    url_id TEXT NOT NULL,
    accessed_at DATETIME NOT NULL,
    ip TEXT,
    FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE
);
