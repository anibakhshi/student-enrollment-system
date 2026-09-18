BEGIN;

DROP TRIGGER IF EXISTS courses_set_updated_at
    ON courses;

DROP TRIGGER IF EXISTS instructors_set_updated_at
    ON instructors;

DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS instructors;

DELETE FROM schema_migrations
WHERE version = '000002';

COMMIT;