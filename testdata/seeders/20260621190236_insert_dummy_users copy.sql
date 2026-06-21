-- Write your data seeding statements here
-- BEST PRACTICE: Make it idempotent so it can be safely executed multiple times.
-- Use "ON CONFLICT DO NOTHING" or "ON CONFLICT (...) DO UPDATE".

-- Example:
INSERT INTO users (id, username) 
VALUES (1, 'superadmin') 
ON CONFLICT (id) DO NOTHING;

