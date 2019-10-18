package migrations

import (
	"fmt"

	"code.hackerspace.pl/hscloud/go/mirko"
	"github.com/golang-migrate/migrate/v4"
)

func New(dburl string) (*migrate.Migrate, error) {
	source, err := mirko.NewMigrationsFromBazel(Data)
	if err != nil {
		return nil, fmt.Errorf("could not create migrations: %v", err)
	}
	return migrate.NewWithSourceInstance("bazel", source, dburl)
}
