ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';

-- На уже живущем узле владельцем становится самый ранний аккаунт: узел
-- поднимает один человек, и это он. Свежая база остаётся пустой — первого
-- зарегистрированного владельцем назначает register.
UPDATE users SET role = 'owner'
WHERE uuid = (SELECT uuid FROM users ORDER BY created_at, uuid LIMIT 1);
