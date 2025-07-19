DROP INDEX IF EXISTS idx_records_user_date;
DROP INDEX IF EXISTS idx_records_reservation;
DROP INDEX IF EXISTS idx_records_user_id;
DROP INDEX IF EXISTS idx_records_date;
DROP INDEX IF EXISTS idx_records_deleted_at;
DROP INDEX IF EXISTS idx_records_user_date_desc;

DROP INDEX IF EXISTS idx_reservations_user_date;
DROP INDEX IF EXISTS idx_reservations_timeslot;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_user_id;
DROP INDEX IF EXISTS idx_reservations_date;
DROP INDEX IF EXISTS idx_reservations_deleted_at;
DROP INDEX IF EXISTS idx_reservations_user_date_desc;

DROP INDEX IF EXISTS idx_users_openid;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_deleted_at;

DROP TABLE IF EXISTS records;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS users; 