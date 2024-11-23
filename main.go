package main

import (
	"embed"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/x0k/effective-mobile-song-library-service/app"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func main() {
	source, _ := iofs.New(migrations, "db/migrations")
	app.Run(source)
}
