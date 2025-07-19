-- 为现有用户设置默认电价
UPDATE users SET unit_price = 0.7 WHERE unit_price IS NULL OR unit_price = 0;

-- 设置默认值约束
ALTER TABLE users ALTER COLUMN unit_price SET DEFAULT 0.7; 