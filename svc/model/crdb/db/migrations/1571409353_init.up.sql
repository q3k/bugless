-- Copyright 2020 Sergiusz Bazanski <q3k@q3k.org>
-- SPDX-License-Identifier: AGPL-3.0-or-later

-- Tree of issue categories.
CREATE TABLE categories (
    -- Opaque identifier.
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Identifier of parent, or NULL if _this_ element is the root of the category tree.
    parent_id UUID,

    -- Human-readable category name, unique among siblings.
    -- Unqiue to avoid confusion when navigatin trees of categories.
    name STRING NOT NULL,

    -- Human-readable description.
    description STRING NOT NULL,

    CONSTRAINT fk_parent FOREIGN KEY (parent_id) REFERENCES categories (id), 
    UNIQUE (parent_id, name)
);

-- Root element, always with a zero UUID.
-- It's also the only element that has a NULL parent_id (ensured by application
-- code).
INSERT INTO categories
    (id, parent_id, name, description)
VALUES (
    '00000000-0000-0000-0000-000000000000',
    NULL,
    'root',
    'Root Category'
);
