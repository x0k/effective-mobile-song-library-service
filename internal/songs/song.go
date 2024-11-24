package songs

import "time"

type Song struct {
	ID          int64
	Title       string
	Artist      string
	ReleaseDate time.Time
	Lyrics      []string
	Link        string
}

func NewSong(
	title string,
	artist string,
	releaseDate time.Time,
	lyrics []string,
	link string,
) Song {
	return Song{
		ID:          -1,
		Title:       title,
		Artist:      artist,
		ReleaseDate: releaseDate,
		Lyrics:      lyrics,
		Link:        link,
	}
}
