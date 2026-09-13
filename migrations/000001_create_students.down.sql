BEGIN;

DROP TRIGGER IF EXISTS students_set_updated_at ON students;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS students;

DELETE FROM schema_migrations
WHERE version = '000001';

DROP TABLE IF EXISTS schema_migrations;

COMMIT;