BEGIN;

DROP TRIGGER IF EXISTS payments_set_updated_at ON payments;

DROP TABLE IF EXISTS payments;

DELETE FROM schema_migrations
WHERE version = '000004';

COMMIT;