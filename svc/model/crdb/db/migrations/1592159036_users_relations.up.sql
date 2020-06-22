-- Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
-- SPDX-License-Identifier: AGPL-3.0-or-later

-- This migration is the second step of a two-step migration, the second step
-- being 1592159036.

ALTER TABLE issues
    DROP COLUMN author,
    DROP COLUMN assignee,
    ALTER COLUMN author_id SET NOT NULL,
    ALTER COLUMN assignee_id SET NOT NULL,
    ADD CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES users (id),
    ADD CONSTRAINT fk_assignee FOREIGN KEY (assignee_id) REFERENCES users (id);

ALTER TABLE issue_updates
    DROP COLUMN author,
    DROP COLUMN assignee,
    ALTER COLUMN author_id SET NOT NULL,
    ADD CONSTRAINT fk_assignee FOREIGN KEY (assignee_id) REFERENCES users (id),
    ADD CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES users (id);
