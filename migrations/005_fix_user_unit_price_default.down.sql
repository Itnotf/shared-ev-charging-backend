ALTER TABLE users ALTER COLUMN unit_price DROP DEFAULT;
UPDATE users SET unit_price = NULL WHERE unit_price = 0.7; 