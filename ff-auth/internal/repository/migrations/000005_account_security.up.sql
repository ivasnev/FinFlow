-- Existing profiles are preserved, but are not silently marked as email-verified.
CREATE TABLE account_security (
 user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
 verified_email TEXT NOT NULL,
 verified_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX account_security_email ON account_security(lower(verified_email));
CREATE TABLE account_pending (
 id UUID PRIMARY KEY,
 secret_hash TEXT NOT NULL UNIQUE,
 email TEXT NOT NULL,
 original_email TEXT NOT NULL,
 nickname TEXT NOT NULL,
 name TEXT,
 phone TEXT,
 password_hash TEXT NOT NULL,
 user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
 existing_user BOOLEAN NOT NULL DEFAULT false,
 expires_at TIMESTAMPTZ NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX account_pending_expiry ON account_pending(expires_at);
CREATE INDEX account_pending_email ON account_pending(lower(email));
-- Email tokens are purpose-bound and single-use. Only the outbox contains an encrypted delivery payload.
CREATE TABLE account_email_tokens (
 id UUID PRIMARY KEY,
 token_hash TEXT NOT NULL UNIQUE,
 purpose TEXT NOT NULL CHECK (purpose IN ('verify','reset','email')),
 pending_id UUID REFERENCES account_pending(id) ON DELETE CASCADE,
 user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
 email TEXT NOT NULL,
 password_hash TEXT NOT NULL DEFAULT '',
 expires_at TIMESTAMPTZ NOT NULL,
 used_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX account_email_tokens_expiry ON account_email_tokens(expires_at);
CREATE TABLE account_access (
 token_hash TEXT PRIMARY KEY,
 session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX account_access_expiry ON account_access(expires_at);
-- Fixed windows are atomic across replicas; identities are hashed, not stored in clear text.
CREATE TABLE account_limits (
 key TEXT PRIMARY KEY,
 count BIGINT NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX account_limits_expiry ON account_limits(expires_at);
CREATE TABLE account_captcha (
 id UUID PRIMARY KEY,
 answer_hash TEXT NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX account_captcha_expiry ON account_captcha(expires_at);
CREATE TABLE account_mail_outbox (
 id UUID PRIMARY KEY REFERENCES account_email_tokens(id) ON DELETE CASCADE,
 payload BYTEA NOT NULL,
 attempts INTEGER NOT NULL DEFAULT 0,
 available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 lease_until TIMESTAMPTZ,
 lease_token UUID,
 sent_at TIMESTAMPTZ,
 failed_at TIMESTAMPTZ,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX account_mail_ready ON account_mail_outbox(available_at) WHERE sent_at IS NULL AND failed_at IS NULL;
