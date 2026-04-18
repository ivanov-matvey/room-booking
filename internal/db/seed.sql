BEGIN;

-- CLEANUP
TRUNCATE TABLE bookings CASCADE;
TRUNCATE TABLE slots CASCADE;
TRUNCATE TABLE schedules CASCADE;
TRUNCATE TABLE rooms CASCADE;
DELETE FROM users
WHERE id NOT IN (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002'
);

-- 1. СОТРУДНИКИ: 10 000 человек
INSERT INTO users (id, email, role, password_hash)
SELECT
    gen_random_uuid(),
    lower(first_name) || '.' || lower(last_name) || lpad(i::text, 4, '0') || '@company.ru',
    CASE WHEN i % 200 = 0 THEN 'admin' ELSE 'user' END,
    NULL
FROM generate_series(1, 10000) AS i
CROSS JOIN LATERAL (
    SELECT
        (ARRAY[
            'alexander','mikhail','dmitry','sergey','andrey','ivan','alexey','nikolay',
            'vladimir','pavel','artem','denis','evgeny','roman','ilya','oleg','anton',
            'vitaly','kirill','maxim','anna','maria','elena','natalia','olga','tatiana',
            'irina','svetlana','ekaterina','yulia','victoria','anastasia','marina',
            'daria','alina','valeria','polina','kristina','sofia','artyom'
        ])[ 1 + (i % 40) ] AS first_name,
        (ARRAY[
            'ivanov','petrov','sidorov','kozlov','novikov','morozov','volkov','sokolov',
            'popov','lebedev','fedorov','mikhailov','belyaev','tarasov','nikitin',
            'frolov','zaytsev','pavlov','semyonov','golubev','vinogradov','bogdanov',
            'vorobyov','gusev','stepanov','andreev','baranov','romanov','kovalev',
            'kuznetsov','solovyov','vasiliev','zaitsev','makarov','nikiforov','orlov',
            'medvedev','smirnov','egorov','titov'
        ])[ 1 + ((i / 40) % 40) ] AS last_name
) AS names
ON CONFLICT DO NOTHING;

-- 2. ПЕРЕГОВОРКИ: 50 комнат
INSERT INTO rooms (id, name, description, capacity)
SELECT
    gen_random_uuid(),
    'Корпус ' || chr(64 + b) || ', эт. ' || f || ', ' || room_name,
    room_desc,
    room_cap
FROM generate_series(1, 5) AS b
CROSS JOIN generate_series(1, 1) AS f
CROSS JOIN (VALUES
    (1,  'Переговорная «Меркурий»',   'Малая, 4 места, проектор',        4),
    (2,  'Переговорная «Венера»',     'Малая, 6 мест, ТВ-панель',        6),
    (3,  'Переговорная «Марс»',       'Средняя, 8 мест, видеосвязь',     8),
    (4,  'Переговорная «Юпитер»',     'Средняя, 10 мест, проектор',     10),
    (5,  'Переговорная «Сатурн»',     'Средняя, 12 мест, ТВ-панель',    12),
    (6,  'Переговорная «Уран»',       'Большая, 16 мест, видеостена',   16),
    (7,  'Переговорная «Нептун»',     'Большая, 20 мест, проектор',     20),
    (8,  'Конференц-зал «Орион»',     'Зал, 30 мест, сцена',            30),
    (9,  'Конференц-зал «Андромеда»', 'Зал, 40 мест, синхроперевод',    40),
    (10, 'Зал совещаний «Галактика»', 'Большой зал, 50 мест',           50)
) AS t(rnum, room_name, room_desc, room_cap);

-- 3. РАСПИСАНИЯ: пн–вс, 08:00–18:00
-- 50 комнат × 20 слотов/день = 1 000 слотов/день
INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time)
SELECT gen_random_uuid(), r.id, ARRAY[1,2,3,4,5,6,7], '08:00', '18:00'
FROM rooms r
LEFT JOIN schedules s ON s.room_id = r.id
WHERE s.id IS NULL;

-- 4. СЛОТЫ: 115 прошлых дней + 90 будущих дней, 20 слотов/день/комната
-- 50 комнат × 206 дней × 20 слотов = ~206 000 слотов
INSERT INTO slots (id, room_id, start_time, end_time)
SELECT
    gen_random_uuid(),
    s.room_id,
    (d + (n * interval '30 minutes')),
    (d + ((n + 1) * interval '30 minutes'))
FROM schedules s
CROSS JOIN generate_series(
    (current_date - interval '115 days')::timestamptz + interval '8 hours',
    (current_date + interval '90 days')::timestamptz  + interval '8 hours',
    interval '1 day'
) AS d
CROSS JOIN generate_series(0, 19) AS n   -- 20 слотов: 08:00–18:00
ON CONFLICT (room_id, start_time) DO NOTHING;

-- 5. БРОНИРОВАНИЯ — реалистичное распределение (~100k)
--
-- Прошлые слоты (start < now):
--   рабочее время (пн-пт 9-17): 70% active
--   нерабочее время:            30% active
--   10% всех прошлых слотов:    cancelled
--
-- Будущие слоты (start >= now):
--   ближайшие 30 дней:  40% active, 8% cancelled
--   31-90 дней:         15% active, 3% cancelled
--
-- Итого: ~64k прошлых active + ~11k cancelled + ~22k future = ~97k броней
CREATE TEMP TABLE tmp_users AS
SELECT id AS user_id, (row_number() OVER (ORDER BY id) - 1) AS rn
FROM users
WHERE role = 'user';

CREATE TEMP TABLE tmp_user_count AS
SELECT COUNT(*)::int AS cnt FROM tmp_users;

-- 5a. Прошлые слоты: активные брони, рабочее время
INSERT INTO bookings (id, slot_id, user_id, status)
SELECT
    gen_random_uuid(),
    s.id,
    (SELECT user_id FROM tmp_users
     WHERE rn = (abs(hashtext(s.id::text)) % (SELECT cnt FROM tmp_user_count))
     LIMIT 1),
    'active'
FROM slots s
WHERE s.start_time < NOW()
  AND EXTRACT(DOW FROM s.start_time) BETWEEN 1 AND 5
  AND EXTRACT(HOUR FROM s.start_time) BETWEEN 9 AND 17
  AND (abs(hashtext(s.id::text)) % 100) < 70
ON CONFLICT DO NOTHING;

-- 5b. Прошлые слоты: активные брони, нерабочее время
INSERT INTO bookings (id, slot_id, user_id, status)
SELECT
    gen_random_uuid(),
    s.id,
    (SELECT user_id FROM tmp_users
     WHERE rn = (abs(hashtext(s.id::text)) % (SELECT cnt FROM tmp_user_count))
     LIMIT 1),
    'active'
FROM slots s
WHERE s.start_time < NOW()
  AND NOT (
      EXTRACT(DOW FROM s.start_time) BETWEEN 1 AND 5
      AND EXTRACT(HOUR FROM s.start_time) BETWEEN 9 AND 17
  )
  AND (abs(hashtext(s.id::text)) % 100) < 30
ON CONFLICT DO NOTHING;

-- 5c. Прошлые слоты: отменённые (10%)
INSERT INTO bookings (id, slot_id, user_id, status)
SELECT
    gen_random_uuid(),
    s.id,
    (SELECT user_id FROM tmp_users
     WHERE rn = ((abs(hashtext(s.id::text)) + 7777) % (SELECT cnt FROM tmp_user_count))
     LIMIT 1),
    'cancelled'
FROM slots s
WHERE s.start_time < NOW()
  AND (abs(hashtext(s.id::text)) % 100) < 10
ON CONFLICT DO NOTHING;

-- 5d. Будущие слоты 0–30 дней: 40% active, 8% cancelled
INSERT INTO bookings (id, slot_id, user_id, status)
SELECT
    gen_random_uuid(),
    s.id,
    (SELECT user_id FROM tmp_users
     WHERE rn = (abs(hashtext(s.id::text)) % (SELECT cnt FROM tmp_user_count))
     LIMIT 1),
    'active'
FROM slots s
WHERE s.start_time >= NOW()
  AND s.start_time < NOW() + interval '30 days'
  AND (abs(hashtext(s.id::text)) % 100) < 40
ON CONFLICT DO NOTHING;

INSERT INTO bookings (id, slot_id, user_id, status)
SELECT
    gen_random_uuid(),
    s.id,
    (SELECT user_id FROM tmp_users
     WHERE rn = ((abs(hashtext(s.id::text)) + 3333) % (SELECT cnt FROM tmp_user_count))
     LIMIT 1),
    'cancelled'
FROM slots s
WHERE s.start_time >= NOW()
  AND s.start_time < NOW() + interval '30 days'
  AND (abs(hashtext(s.id::text)) % 100) BETWEEN 40 AND 47
ON CONFLICT DO NOTHING;

-- 5e. Будущие слоты 30–90 дней: 15% active, 3% cancelled
INSERT INTO bookings (id, slot_id, user_id, status)
SELECT
    gen_random_uuid(),
    s.id,
    (SELECT user_id FROM tmp_users
     WHERE rn = (abs(hashtext(s.id::text)) % (SELECT cnt FROM tmp_user_count))
     LIMIT 1),
    'active'
FROM slots s
WHERE s.start_time >= NOW() + interval '30 days'
  AND (abs(hashtext(s.id::text)) % 100) < 15
ON CONFLICT DO NOTHING;

INSERT INTO bookings (id, slot_id, user_id, status)
SELECT
    gen_random_uuid(),
    s.id,
    (SELECT user_id FROM tmp_users
     WHERE rn = ((abs(hashtext(s.id::text)) + 1111) % (SELECT cnt FROM tmp_user_count))
     LIMIT 1),
    'cancelled'
FROM slots s
WHERE s.start_time >= NOW() + interval '30 days'
  AND (abs(hashtext(s.id::text)) % 100) BETWEEN 15 AND 17
ON CONFLICT DO NOTHING;

-- ФИНАЛЬНАЯ СТАТИСТИКА
SELECT
    (SELECT COUNT(*) FROM users)                                    AS total_users,
    (SELECT COUNT(*) FROM users WHERE role = 'admin')               AS admins,
    (SELECT COUNT(*) FROM users WHERE role = 'user')                AS employees,
    (SELECT COUNT(*) FROM rooms)                                    AS total_rooms,
    (SELECT COUNT(*) FROM schedules)                                AS total_schedules,
    (SELECT COUNT(*) FROM slots)                                    AS total_slots,
    (SELECT COUNT(*) FROM slots WHERE start_time < NOW())           AS past_slots,
    (SELECT COUNT(*) FROM slots WHERE start_time >= NOW())          AS future_slots,
    (SELECT COUNT(*) FROM bookings)                                 AS total_bookings,
    (SELECT COUNT(*) FROM bookings WHERE status = 'active')         AS active_bookings,
    (SELECT COUNT(*) FROM bookings WHERE status = 'cancelled')      AS cancelled_bookings;

COMMIT;
