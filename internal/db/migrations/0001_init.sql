CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE COLLATE NOCASE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE TABLE sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE magic_tokens (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL COLLATE NOCASE,
  token_hash BLOB NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE items (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT 'general',
  per_night INTEGER NOT NULL DEFAULT 0,
  default_qty INTEGER NOT NULL DEFAULT 1,
  notes TEXT,
  created_by TEXT REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted_by TEXT REFERENCES users(id)
);
CREATE INDEX idx_items_active ON items(deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE bundles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT,
  created_by TEXT REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted_by TEXT REFERENCES users(id)
);
CREATE INDEX idx_bundles_active ON bundles(deleted_at) WHERE deleted_at IS NULL;

CREATE TABLE bundle_items (
  bundle_id TEXT NOT NULL REFERENCES bundles(id),
  item_id   TEXT NOT NULL REFERENCES items(id),
  qty       INTEGER,
  PRIMARY KEY (bundle_id, item_id)
);

CREATE TABLE bundle_children (
  parent_id TEXT NOT NULL REFERENCES bundles(id),
  child_id  TEXT NOT NULL REFERENCES bundles(id),
  PRIMARY KEY (parent_id, child_id),
  CHECK (parent_id <> child_id)
);
CREATE INDEX idx_bundle_children_child ON bundle_children(child_id);

CREATE TABLE trips (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  nights INTEGER NOT NULL DEFAULT 1 CHECK (nights >= 0),
  starts_on DATE,
  notes TEXT,
  owner_id TEXT NOT NULL REFERENCES users(id),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP,
  deleted_by TEXT REFERENCES users(id)
);
CREATE INDEX idx_trips_owner_active ON trips(owner_id) WHERE deleted_at IS NULL;

CREATE TABLE trip_members (
  trip_id TEXT NOT NULL REFERENCES trips(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  role TEXT NOT NULL CHECK (role IN ('owner','editor')),
  added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (trip_id, user_id)
);

CREATE TABLE trip_bundles (
  trip_id   TEXT NOT NULL REFERENCES trips(id),
  bundle_id TEXT NOT NULL REFERENCES bundles(id),
  added_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (trip_id, bundle_id)
);

CREATE TABLE trip_extras (
  trip_id TEXT NOT NULL REFERENCES trips(id),
  item_id TEXT NOT NULL REFERENCES items(id),
  qty     INTEGER,
  PRIMARY KEY (trip_id, item_id)
);

CREATE TABLE trip_overrides (
  trip_id      TEXT NOT NULL REFERENCES trips(id),
  item_id      TEXT NOT NULL REFERENCES items(id),
  removed      INTEGER NOT NULL DEFAULT 0,
  qty_override INTEGER,
  PRIMARY KEY (trip_id, item_id)
);

CREATE TABLE trip_pack_state (
  trip_id   TEXT NOT NULL REFERENCES trips(id),
  item_id   TEXT NOT NULL REFERENCES items(id),
  packed    INTEGER NOT NULL DEFAULT 0,
  packed_at TIMESTAMP,
  PRIMARY KEY (trip_id, item_id)
);

