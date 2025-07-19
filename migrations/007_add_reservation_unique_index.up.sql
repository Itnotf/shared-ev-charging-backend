-- 限制同一用户同一天同一时段只能有一条有效预约（不含cancelled）
CREATE UNIQUE INDEX IF NOT EXISTS uniq_reservation_user_date_timeslot ON reservations(user_id, date, timeslot) WHERE status != 'cancelled'; 