CREATE TABLE
  song (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    release_date DATE NOT NULL,
    lyrics TEXT[] NOT NULL,
    link VARCHAR(255) NOT NULL
  );

-- Multiple indexes take up a lot of space and degrade the insertion speed
-- Ideally, indexes should be added after analyzing the data and scenarios for their use

CREATE INDEX idx_song_title ON song (title);
CREATE INDEX idx_song_artist ON song (artist);
CREATE INDEX idx_song_release_date ON song (release_date);
CREATE INDEX idx_song_lyrics ON song USING GIN (lyrics);
