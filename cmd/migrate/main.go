package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/reviveos/utils/db"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	migrationsDir := filepath.Join(".", "database", "migrations")
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var upFiles []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".sql" && len(f.Name()) > 7 && f.Name()[len(f.Name())-7:] == ".up.sql" {
			upFiles = append(upFiles, f.Name())
		}
	}
	sort.Strings(upFiles)

	for _, file := range upFiles {
		filePath := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration %s: %v", file, err)
		}

		log.Printf("Executing migration: %s", file)
		_, err = pool.Exec(ctx, string(content))
		if err != nil {
			log.Fatalf("Failed executing %s: %v", file, err)
		}
		log.Printf("Successfully applied %s", file)
	}

	fmt.Println("All migrations applied successfully!")
}
