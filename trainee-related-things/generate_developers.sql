INSERT INTO developers (name, department, geolocation, last_known_ip, is_available)
SELECT
    (ARRAY['James', 'Mary', 'John', 'Patricia', 'Robert'])[1 + floor(random() * 5)] || ' ' || (ARRAY['Smith', 'Johnson', 'Williams', 'Brown', 'Jones'])[1 + floor(random() * 5)],
    (ARRAY['backend', 'frontend', 'ios', 'android'])[1 + floor(random() * 4)]::department_type,
    point((random() * 360) - 180, (random() * 180) - 90),
    (floor(random() * 255 + 1) || '.' || floor(random() * 256) || '.' || floor(random() * 256) || '.' || floor(random() * 256))::INET,
    random() > 0.5
FROM generate_series(1, 5432);

