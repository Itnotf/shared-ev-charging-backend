-- 移除车牌号外键约束，改为逻辑关联
-- 这样可以提高性能，增加灵活性，避免外键约束的复杂性

-- 移除预约表的外键约束
ALTER TABLE reservations DROP CONSTRAINT IF EXISTS fk_reservations_license_plate_id;

-- 移除记录表的外键约束
ALTER TABLE records DROP CONSTRAINT IF EXISTS fk_records_license_plate_id;

-- 添加注释说明这是逻辑关联
COMMENT ON COLUMN reservations.license_plate_id IS '关联的车牌号ID（逻辑关联，无外键约束）';
COMMENT ON COLUMN records.license_plate_id IS '关联的车牌号ID（逻辑关联，无外键约束）';
