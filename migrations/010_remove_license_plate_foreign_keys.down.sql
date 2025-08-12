-- 恢复车牌号外键约束
-- 重新添加外键约束（如果需要的话）

-- 为预约表重新添加外键约束
ALTER TABLE reservations ADD CONSTRAINT fk_reservations_license_plate_id FOREIGN KEY (license_plate_id) REFERENCES license_plates(id) ON DELETE SET NULL;

-- 为记录表重新添加外键约束
ALTER TABLE records ADD CONSTRAINT fk_records_license_plate_id FOREIGN KEY (license_plate_id) REFERENCES license_plates(id) ON DELETE SET NULL;
