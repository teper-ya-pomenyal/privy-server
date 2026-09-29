-- Блокировка аккаунта владельцем узла (админка, экран «Пользователи»).
ALTER TABLE users ADD COLUMN blocked BOOLEAN NOT NULL DEFAULT false;
