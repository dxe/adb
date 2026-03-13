INSERT INTO users_roles (user_id, role)
SELECT DISTINCT source.user_id, 'non-sfbay'
FROM users_roles source
LEFT JOIN users_roles existing ON existing.user_id = source.user_id AND existing.role = 'non-sfbay'
WHERE source.role = 'organizer' AND existing.user_id IS NULL;
