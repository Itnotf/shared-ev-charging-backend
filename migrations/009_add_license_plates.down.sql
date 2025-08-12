-- 删除外键约束和新增字段
ALTER TABLE records DROP CONSTRAINT IF EXISTS fk_records_license_plate_id;
ALTER TABLE records DROP COLUMN IF EXISTS license_plate_id;

ALTER TABLE reservations DROP CONSTRAINT IF EXISTS fk_reservations_license_plate_id;
ALTER TABLE reservations DROP COLUMN IF EXISTS license_plate_id;

-- 删除车牌号表
DROP TABLE IF EXISTS license_plates;


