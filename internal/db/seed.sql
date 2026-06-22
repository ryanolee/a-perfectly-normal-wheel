PRAGMA foreign_keys = OFF;
DELETE FROM candidates;
DELETE FROM wheels;
PRAGMA foreign_keys = ON;

INSERT INTO wheels (id, name, description) VALUES
    (1, 'Weekly Standup', 'Who goes first in standup this week?'),
    (2, 'Code Review', 'Pick a reviewer for the latest PR');

INSERT INTO candidates (username, wheel_id, creator_id) VALUES
    ('alice', 1, 'seed'),
    ('bob', 1, 'seed'),
    ('carol', 1, 'seed'),
    ('dave', 2, 'seed'),
    ('eve', 2, 'seed');
