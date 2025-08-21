package main

import (
	"os"
	"shortleak/database"
	packages_migrate "shortleak/packages/migrate"
)

var newMigrator = func() database.Migrator {
	return &database.DefaultMigrator{}
}

func RunMigrate(args []string, m database.Migrator) {
	packages_migrate.Run(args, m)
}

func main() {
	m := newMigrator()
	RunMigrate(os.Args[1:], m)
}
