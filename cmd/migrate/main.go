package main

import (
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dsn := flag.String("dsn", "", "MySQL DSN")
	dir := flag.String("dir", "db/migrations", "migrations directory")
	flag.Parse()

	m, err := migrate.New(
		"file://"+*dir,
		"mysql://"+*dsn,
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}
