DROP INDEX IF EXISTS idx_tracks_artist_listened;

ALTER TABLE tracks
    ALTER COLUMN listened DROP NOT NULL,
    ALTER COLUMN listened DROP DEFAULT;
