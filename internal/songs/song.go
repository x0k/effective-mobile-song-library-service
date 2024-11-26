package songs

import (
	"time"
)

const releaseDateFormat = "02.01.2006"

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

type Pagination struct {
	PageSize uint64
	Page     uint64
}

type Query struct {
	Pagination
	LastId int64
	Filter string
}
