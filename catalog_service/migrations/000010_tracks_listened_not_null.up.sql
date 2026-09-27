UPDATE tracks SET listened = 0 WHERE listened IS NULL;

ALTER TABLE tracks
    ALTER COLUMN listened SET DEFAULT 0,
    ALTER COLUMN listened SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_tracks_artist_listened
ON tracks (artist_id, listened DESC);
