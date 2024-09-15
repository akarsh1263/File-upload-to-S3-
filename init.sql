CREATE TABLE users (
    email VARCHAR(255) PRIMARY KEY,
    password TEXT NOT NULL
);

CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) REFERENCES users(email),
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    upload_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    file_size INTEGER,
    file_type VARCHAR(255) NOT NULL
);

CREATE INDEX idx_file_name ON files(file_name);
CREATE INDEX idx_upload_date ON files(upload_date);
CREATE INDEX idx_file_type ON files(file_type);
