-- 车牌号表
CREATE TABLE IF NOT EXISTS license_plates (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    plate_number VARCHAR(20) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_license_plates_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 为预约表添加车牌号字段
ALTER TABLE reservations ADD COLUMN license_plate_id INTEGER;
ALTER TABLE reservations ADD CONSTRAINT fk_reservations_license_plate_id 
    FOREIGN KEY (license_plate_id) REFERENCES license_plates(id) ON DELETE SET NULL;

-- 为记录表添加车牌号字段
ALTER TABLE records ADD COLUMN license_plate_id INTEGER;
ALTER TABLE records ADD CONSTRAINT fk_records_license_plate_id 
    FOREIGN KEY (license_plate_id) REFERENCES license_plates(id) ON DELETE SET NULL;

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_license_plates_user_id ON license_plates(user_id);
CREATE INDEX IF NOT EXISTS idx_license_plates_deleted_at ON license_plates(deleted_at);
CREATE INDEX IF NOT EXISTS idx_license_plates_plate_number ON license_plates(plate_number);
CREATE UNIQUE INDEX IF NOT EXISTS idx_license_plates_user_default 
    ON license_plates(user_id) WHERE is_default = TRUE AND deleted_at IS NULL;

-- 为新字段创建索引
CREATE INDEX IF NOT EXISTS idx_reservations_license_plate_id ON reservations(license_plate_id);
CREATE INDEX IF NOT EXISTS idx_records_license_plate_id ON records(license_plate_id);

-- 添加注释
COMMENT ON TABLE license_plates IS '用户车牌号表';
COMMENT ON COLUMN license_plates.user_id IS '用户ID';
COMMENT ON COLUMN license_plates.plate_number IS '车牌号';
COMMENT ON COLUMN license_plates.is_default IS '是否为默认车牌号';
COMMENT ON COLUMN reservations.license_plate_id IS '关联的车牌号ID';
COMMENT ON COLUMN records.license_plate_id IS '关联的车牌号ID';


