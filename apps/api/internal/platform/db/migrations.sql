CREATE TABLE IF NOT EXISTS admin_users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS app_settings (
    id SMALLINT PRIMARY KEY,
    default_provider TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS source_assets (
    id TEXT PRIMARY KEY,
    asset_type TEXT NOT NULL,
    title TEXT NOT NULL,
    file_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    binary_data BYTEA NOT NULL,
    extraction_status TEXT NOT NULL,
    extracted_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS application_specs (
    id TEXT PRIMARY KEY,
    company_name TEXT NOT NULL,
    role_name TEXT NOT NULL,
    source_text TEXT NOT NULL DEFAULT '',
    warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
    questions JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS writing_sessions (
    id TEXT PRIMARY KEY,
    application_spec_id TEXT NOT NULL REFERENCES application_specs(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    current_provider TEXT NOT NULL,
    apply_mode TEXT NOT NULL,
    review_mode TEXT NOT NULL,
    auto_review BOOLEAN NOT NULL DEFAULT TRUE,
    auto_apply BOOLEAN NOT NULL DEFAULT FALSE,
    count_mode TEXT NOT NULL,
    finalized_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS session_assets (
    session_id TEXT NOT NULL REFERENCES writing_sessions(id) ON DELETE CASCADE,
    asset_id TEXT NOT NULL REFERENCES source_assets(id) ON DELETE CASCADE,
    PRIMARY KEY (session_id, asset_id)
);

CREATE TABLE IF NOT EXISTS question_drafts (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES writing_sessions(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL,
    title TEXT NOT NULL,
    prompt_text TEXT NOT NULL,
    char_limit INTEGER NOT NULL DEFAULT 0,
    document_json JSONB NOT NULL,
    plain_text TEXT NOT NULL DEFAULT '',
    inferred_count INTEGER NOT NULL DEFAULT 0,
    resolved_inferred_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, question_id)
);

CREATE TABLE IF NOT EXISTS session_versions (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES writing_sessions(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    document_json JSONB NOT NULL,
    plain_text TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES writing_sessions(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL,
    message_type TEXT NOT NULL,
    content TEXT NOT NULL,
    meta_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tool_actions (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES writing_sessions(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL DEFAULT '',
    tool_name TEXT NOT NULL,
    prompt_text TEXT NOT NULL,
    chat_message_id TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS suggestions (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES writing_sessions(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL,
    source TEXT NOT NULL,
    scope TEXT NOT NULL,
    rationale JSONB NOT NULL DEFAULT '[]'::jsonb,
    original_document_json JSONB NOT NULL,
    original_plain_text TEXT NOT NULL DEFAULT '',
    suggested_document_json JSONB NOT NULL,
    suggested_plain_text TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    applied_at TIMESTAMPTZ NULL
);

CREATE TABLE IF NOT EXISTS review_reports (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES writing_sessions(id) ON DELETE CASCADE,
    question_id TEXT NOT NULL DEFAULT '',
    report_json JSONB NOT NULL,
    ready_to_submit BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
