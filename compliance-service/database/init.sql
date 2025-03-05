-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_name TEXT UNIQUE NOT NULL,
    secret_code TEXT NOT NULL
);

-- Create cards table
CREATE TABLE IF NOT EXISTS cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    card_number TEXT UNIQUE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Create reported_cards table
CREATE TABLE IF NOT EXISTS reported_cards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    card_id INTEGER NOT NULL,
    reported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (card_id) REFERENCES cards (id) ON DELETE CASCADE,
    UNIQUE (user_id, card_id)
);

-- DUMMY DATA
INSERT OR IGNORE INTO users (user_name, secret_code) VALUES 
    ('john_doe', '$2a$10$0cdvmI6GCiqRozURednLDOX0wyWHx9HYOOjQhmdFXOSSKYYUC7Oca'),   -- secret_code: hashed_secret_123
    ('jane_smith', '$2a$10$xe4/MMDeSW5Qj59sXAriS.3tMjMPzlQh6MX/Qr2frrNggCiI.29ZO'); -- secret_code: hashed_secret_456

INSERT OR IGNORE INTO cards (user_id, card_number) VALUES 
    (1, '1234-5678-9012-3456'),
    (1, '9876-5432-1098-7654'),
    (2, '1122-3344-5566-7788');
