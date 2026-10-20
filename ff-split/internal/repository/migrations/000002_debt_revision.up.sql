ALTER TABLE events ADD COLUMN debt_revision BIGINT NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN optimized_debt_revision BIGINT;
