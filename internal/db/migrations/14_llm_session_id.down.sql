DROP INDEX IF EXISTS idx_llm_conversations_session_id;

ALTER TABLE llm_conversations DROP COLUMN session_id;
