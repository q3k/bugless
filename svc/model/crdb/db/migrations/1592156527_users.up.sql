-- Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
-- SPDX-License-Identifier: AGPL-3.0-or-later

-- This migration adds the 'users' table and changes all user columns in
-- existing tables to be a foreign key to the new table.
-- If this was a real migration in a released version of Bugless, it would've
-- been multi-stage, in lockstep with the application to create user objects.
-- However, this is a lot of wasted effort since Bugless hasn't haf a single
-- release yet (and as per the README at the time of writing, no stability is
-- guaranteed), so we just drop all data instead
-- To reiterate in caps, THIS EATS YOUR DATA. If for some reason you're already
-- holding important data in Bugless - god help you.

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

-- Nuke existing data, as per header.
DELETE FROM issues;
DELETE FROM issue_updates;

ALTER TABLE issues
    DROP COLUMN author,
    DROP COLUMN assignee,
    ADD COLUMN author_id UUID NOT NULL,
    ADD COLUMN assignee_id UUID NOT NULL,
    ADD CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES users (id),
    ADD CONSTRAINT fk_assignee FOREIGN KEY (assignee_id) REFERENCES users (id);

ALTER TABLE issue_updates
    DROP COLUMN author,
    DROP COLUMN assignee,
    ADD COLUMN author_id UUID NOT NULL,
    ADD COLUMN assignee_id UUID,
    ADD CONSTRAINT fk_assignee FOREIGN KEY (assignee_id) REFERENCES users (id),
    ADD CONSTRAINT fk_author FOREIGN KEY (author_id) REFERENCES users (id);

-- issue_{update_,}cc_lists are unused by code at this point - just nuke all
-- and recreate.
DROP TABLE issue_cc_lists;
DROP TABLE issue_update_cc_lists;

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
