-- Refuse ambiguous existing identities instead of silently merging user data.
CREATE UNIQUE INDEX users_email_fold_key ON users(lower(email));
