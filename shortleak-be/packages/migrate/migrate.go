package packages_migrate

import (
	"fmt"
	"shortleak/config"
	"shortleak/database"
)

func Run(args []string, migrator database.Migrator) {
	cfg := config.LoadConfig()
	database.ConnectDB(cfg)

	if len(args) < 1 {
		fmt.Println("Usage: go run main.go [migrate|rollback|refresh]")
		return
	}

	cmd := args[0]

	switch cmd {
	case "migrate":
		if err := migrator.Migrate(database.DB); err != nil {
			fmt.Println("❌ Migration failed:", err)
		} else {
			fmt.Println("✅ Migration success!")
		}
	case "rollback":
		if err := migrator.RollbackAll(database.DB); err != nil {
			fmt.Println("❌ Rollback failed:", err)
		} else {
			fmt.Println("✅ Rollback success!")
		}
	case "refresh":
		if migrator.HasMigrationsTable() {
			if err := migrator.RollbackAll(database.DB); err != nil {
				fmt.Println("⚠️ Rollback failed:", err)
			}
		}
		if err := migrator.Migrate(database.DB); err != nil {
			fmt.Println("❌ Migration failed:", err)
		} else {
			fmt.Println("✅ Migration success!")
		}
		if err := migrator.Seed(); err != nil {
			fmt.Println("❌ Seeding failed:", err)
		} else {
			fmt.Println("✅ Seeding success!")
		}
	default:
		fmt.Println("Unknown command:", cmd)
	}
}
