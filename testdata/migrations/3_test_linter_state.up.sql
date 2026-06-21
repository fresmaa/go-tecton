-- First query is valid but will be blocked by linter
ALTER TABLE users ADD COLUMN email VARCHAR(100) DEFAULT 'admin@domain.com';

-- Second query is invalid and will cause an error
DROP TABLE dummy_table;