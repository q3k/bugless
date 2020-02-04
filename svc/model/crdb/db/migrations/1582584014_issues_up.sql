-- Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Issue numbers. This is used as the single, numerical namespace for all
-- created issues.
CREATE SEQUENCE issue_numbers NO CYCLE;

-- Issues.
CREATE TABLE issues (
    -- The issue_numbers sequence is used to feed the primary key of issues.
    -- Alternatively, we could use an UUID as a primary key and a secondary
    -- flat number for the external key, but that probably doesn't make much
    -- practical sense.
    id SERIAL8 PRIMARY KEY DEFAULT nextval('issue_numbers'),

    --- The following fields never change.
    -- Who reported the bug. Opaque authn string.
    reporter STRING NOT NULL,
    -- When the issue was created, int64 nanos since epoch.
    created INT NOT NULL,

    -- Issues are a historical log of changes. Instead of doing the usual
    -- long-then-compact appraoch, we're keeping a pure log in a different
    -- table (as it will be rendered by users), and a denormalized compacted
    -- view here. All changes performed to these rows must be done in the same
    -- transaction that appends to the log. Two concurrent append/edit
    -- transactions will fence eachother off at the update of this table.
    -- Since CrDB does not provide an explicit row-locking mechanism, we
    -- lock on updating this table's last_modified timestamp.

    -- Last update of the issue, int64 nanos since epoch.
    last_updated INT NOT NULL,

    --- Denormalized, compacted data. Again, only update this within a
    --- transaction that appends to issue_updates and also bumps last_updated.

    -- One-line title of the issue.
    title STRING NOT NULL,
    -- Opaque authn string of who the bug is assigned to. Empty string if
    -- no-one.
    assignee STRING NOT NULL,

    -- Standard fields. These have a fixed schema in svc/model/common.proto.

    -- Issue type. Synchronized to bugless.svc.model.common.IssueType.
    "type" INT8 check (
        "type" >= 0 and "type" <= 6
    ) NOT NULL,
    -- Issue priority, P0 to P4.
    priority INT8 check (
        priority >= 0 and priority <= 4
    ) NOT NULL,
    -- Issue status. Synchronized to bugless.svc.model.common.IssueStatus.
    status INT8 check (
        status >= 0 and status <= 11
    ) NOT NULL
);

-- Issue CC lists.
CREATE TABLE issue_cc_lists (
    issue_id INT8 NOT NULL,

    -- The actual member of the CC list. Opaque authn string.
    member STRING NOT NULL,

    PRIMARY KEY (issue_id, member),
    CONSTRAINT fk_issue FOREIGN KEY (issue_id) REFERENCES issues (id),
    UNIQUE (issue_id, member)
) INTERLEAVE IN PARENT issues (issue_id);

-- Issue update log.
-- This is what makes this an issue tracker - it tracks updates to issues
-- in a separate table.
-- This is append-only and all appends must also update the parent issue. See
-- the comment about last_updated in issues.
CREATE TABLE issue_updates(
    issue_id INT8 NOT NULL,

    -- Updates just use an UUID. In the interface they also have a numerical
    -- ID starting at #1 within an issue, but this is application generated.
    id UUID DEFAULT gen_random_uuid(),

    -- When this update was created, int64 nanos since epoch.
    created INT NOT NULL,
    -- Who created this update. Opaque authn string.
    author STRING NOT NULL,

    --- In updates, all update'eable fields are nullable. A null column
    --- indicates no update. These have the same meaning as in the issues
    --- table.

    title STRING NOT NULL,
    assignee STRING NOT NULL,
    "type" INT8 check (
        "type" >= 0 and "type" <= 6
    ) NOT NULL,
    priority INT8 check (
        priority >= 0 and priority <= 4
    ) NOT NULL,
    status INT8 check (
        status >= 0 and status <= 11
    ) NOT NULL,

    PRIMARY KEY (issue_id, id),
    CONSTRAINT fk_issue FOREIGN KEY (issue_id) REFERENCES issues (id)
);
