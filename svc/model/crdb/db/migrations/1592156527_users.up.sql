-- Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
-- SPDX-License-Identifier: AGPL-3.0-or-later

-- This migration is the first step of a two-step migration, the second step
-- being 1592159036.

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username STRING NOT NULL,

    -- Blob of user preferences proto, //proto/userprefs/userprefs.proto.
    preferences BLOB NOT NULL,

    -- All the following are optionally stored in the databasse. Depending on
    -- the IdP used, they might be always requested from IdP, stored in memory,
    -- cached in the database, or plainly stored in the database.

    -- Email under which user is reachable, for email-based updates.
    email STRING,
    -- Prefered display name for the user when looking at profile information.
    -- This might be a full name, or first name, or anything else.
    display_name STRING,

    UNIQUE (username)
);

-- 'unassigned' user. This is who issues are assigned to when they are
-- unassigned.
-- We use an 'unassigned' user for simplicity of schema of the issue_updates
-- table, to differentiate NULL (no update) from updating-to-unassigned.
INSERT INTO users
    (id, username, preferences)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    '', ''
);

ALTER TABLE issues
    ADD COLUMN author_id UUID,
    ADD COLUMN assignee_id UUID;

ALTER TABLE issue_updates
    ADD COLUMN author_id UUID,
    ADD COLUMN assignee_id UUID;

-- issue_{update_,}cc_lists are unused by code at this point - just nuke all
-- and recreate.
DROP TABLE issue_cc_lists;
DROP TABLE issue_update_cc_lists;

-- Issue CC lists.
CREATE TABLE issue_cc_lists (
    issue_id INT8 NOT NULL,

    -- The actual member of the CC list.
    member_id UUID NOT NULL,

    PRIMARY KEY (issue_id, member_id),
    CONSTRAINT fk_issue FOREIGN KEY (issue_id) REFERENCES issues (id),
    CONSTRAINT fk_member FOREIGN KEY (member_id) REFERENCES users (id)
) INTERLEAVE IN PARENT issues (issue_id);

CREATE TABLE issue_update_cc_lists (
    issue_id INT8 NOT NULL,
    update_id INT8 NOT NULL,

    -- The actual member of the CC list.
    member_id UUID NOT NULL,

    PRIMARY KEY (issue_id, update_id, member_id),
    CONSTRAINT fk_issue_update FOREIGN KEY (issue_id, update_id) REFERENCES issue_updates(issue_id, id),
    CONSTRAINT fk_member FOREIGN KEY (member_id) REFERENCES users (id)
) INTERLEAVE IN PARENT issue_updates (issue_id, update_id);
