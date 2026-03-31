ALTER TABLE llm_conversations ADD COLUMN session_id VARCHAR(64);

CREATE UNIQUE INDEX IF NOT EXISTS idx_llm_conversations_session_id ON llm_conversations(session_id);
