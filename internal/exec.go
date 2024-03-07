package internal

const CreateTableSQL = `
	CREATE SEQUENCE IF NOT EXISTS users_id_seq;
	CREATE SEQUENCE IF NOT EXISTS messages_id_seq;

	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY DEFAULT nextval('users_id_seq'),
		name VARCHAR(200) NOT NULL,
		created_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id BIGINT PRIMARY KEY DEFAULT nextval('messages_id_seq'),
		from_id BIGINT NOT NULL,
		to_id BIGINT NOT NULL,
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
