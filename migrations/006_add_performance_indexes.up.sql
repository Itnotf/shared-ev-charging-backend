-- 添加性能优化相关的数据库索引

-- 充电记录表索引
CREATE INDEX IF NOT EXISTS idx_records_user_date ON records(user_id, date);
CREATE INDEX IF NOT EXISTS idx_records_reservation ON records(reservation_id);
CREATE INDEX IF NOT EXISTS idx_records_user_id ON records(user_id);

-- 预约表索引
CREATE INDEX IF NOT EXISTS idx_reservations_user_date ON reservations(user_id, date);
CREATE INDEX IF NOT EXISTS idx_reservations_timeslot ON reservations(timeslot);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_openid ON users(openid);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- 复合索引优化
CREATE INDEX IF NOT EXISTS idx_records_user_date_desc ON records(user_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_reservations_user_date_desc ON reservations(user_id, date DESC); 