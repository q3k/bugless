-- Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
-- SPDX-License-Identifier: AGPL-3.0-or-later

-- This nukes your data. See header in corresponding .up. migration.

DELETE FROM issues;
DELETE FROM issue_udpates;

ALTER TABLE issues
    DROP CONSTRAINT fk_author,
    DROP CONSTRAINT fk_assignee,
    DROP COLUMN author_id,
    DROP COLUMN assignee_id,
    ADD COLUMN author STRING NOT NULL,
    ADD COLUMN assignee STRING NOT NULL;

ALTER TABLE issue_updates
    DROP CONSTRAINT fk_author,
    DROP CONSTRAINT fk_assignee,
    DROP COLUMN author_id,
    DROP COLUMN assignee_id,
    ADD COLUMN author STRING NOT NULL,
    ADD COLUMN assignee STRING;

DROP TABLE users;
