BEGIN;

DROP TRIGGER IF EXISTS enrollments_set_updated_at
    ON enrollments;

DROP TABLE IF EXISTS enrollments;

DELETE FROM schema_migrations
WHERE version = '000003';

COMMIT;