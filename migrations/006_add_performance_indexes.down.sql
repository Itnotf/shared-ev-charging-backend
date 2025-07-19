-- 回滚性能优化相关的数据库索引

-- 删除充电记录表索引
DROP INDEX IF EXISTS idx_records_user_date;
DROP INDEX IF EXISTS idx_records_reservation;
DROP INDEX IF EXISTS idx_records_user_id;
DROP INDEX IF EXISTS idx_records_user_date_desc;

-- 删除预约表索引
DROP INDEX IF EXISTS idx_reservations_user_date;
DROP INDEX IF EXISTS idx_reservations_timeslot;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_user_date_desc;

-- 删除用户表索引
DROP INDEX IF EXISTS idx_users_openid;
DROP INDEX IF EXISTS idx_users_status; 