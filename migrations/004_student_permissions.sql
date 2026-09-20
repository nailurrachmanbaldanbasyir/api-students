INSERT INTO permissions (name, description) VALUES
    ('student:list', 'melihat daftar seluruh student'),
    ('student:read:any', 'melihat data student mana pun'),
    ('student:create', 'membuat data student'),
    ('student:update:any', 'mengubah data student mana pun'),
    ('student:delete', 'menghapus data student')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
    ('admin', 'student:list'),
    ('admin', 'student:read:any'),
    ('admin', 'student:create'),
    ('admin', 'student:update:any'),
    ('admin', 'student:delete'),
    ('staff', 'student:list'),
    ('staff', 'student:read:any'),
    ('staff', 'student:create')
ON CONFLICT DO NOTHING;

ALTER TABLE students
ADD COLUMN IF NOT EXISTS owner_id INTEGER;

UPDATE students
SET owner_id = 1
WHERE owner_id IS NULL;

ALTER TABLE students
ALTER COLUMN owner_id SET NOT NULL;

ALTER TABLE students
DROP CONSTRAINT IF EXISTS students_owner_id_fkey;

ALTER TABLE students
ADD CONSTRAINT students_owner_id_fkey
FOREIGN KEY (owner_id) REFERENCES users(id);