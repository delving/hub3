BEGIN;
DROP TABLE IF EXISTS namespace;
DROP TABLE IF EXISTS namespace_history;
DROP TRIGGER IF EXISTS namespace_hist_trigger on namespace;
DROP EXTENSION IF EXISTS temporal_tables;
DROP EXTENSION IF EXISTS "uuid-ossp";
END;
