package internal

const CreateTableSQL = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name VARCHAR(200) NOT NULL,
	created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	from_id INTEGER NOT NULL,
	to_id INTEGER NOT NULL,
	content VARCHAR(4096) NOT NULL,
	created_at TIMESTAMP,
	CONSTRAINT fk_messages_from_id FOREIGN KEY (from_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT fk_messages_to_id FOREIGN KEY (to_id) REFERENCES users(id) ON DELETE CASCADE
);
	
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at) WHERE created_at IS NOT NULL; 

CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at) WHERE created_at IS NOT NULL; 
CREATE INDEX IF NOT EXISTS idx_messages_from_id ON messages (from_id); 
CREATE INDEX IF NOT EXISTS idx_messages_to_id ON messages (to_id);
`
