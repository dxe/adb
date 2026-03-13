INSERT INTO users_roles (user_id, role)
SELECT DISTINCT source.user_id, 'organizer'
FROM users_roles source
LEFT JOIN users_roles existing ON existing.user_id = source.user_id AND existing.role = 'organizer'
WHERE source.role = 'non-sfbay' AND existing.user_id IS NULL;

DELETE FROM users_roles
WHERE role = 'non-sfbay';
