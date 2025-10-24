// internal/db/db.go
// This file manages the SQLite database connection and operations using GORM.

package db

import (
	"log"
	"nixon/internal/config"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// Recording represents a recorded audio file.
type Recording struct {
	gorm.Model
	Filename  string `gorm:"uniqueIndex"`
	StartTime time.Time
	Duration  time.Duration
	SizeMB    float64
	Notes     string
	Genre     string
	// The fields below are not currently being used, but are reserved for future features:
	// Tags      string // Comma-separated list of tags
	// OwnerID   uint   // For future security/multi-user features (Phase 3)
}

// InitializeDB sets up the database connection and auto-migrates the schema.
func InitializeDB() {
	dbPath := filepath.Join(config.GetConfig().AutoRecord.Directory, "nixon.db")

	// Ensure the directory exists before attempting to open the DB
	if err := os.MkdirAll(config.GetConfig().AutoRecord.Directory, 0755); err != nil {
		log.Fatalf("FATAL: Could not create database directory: %v", err)
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		// Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{...}) // Optionally enable GORM logging here
	})
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&Recording{}); err != nil {
		log.Fatalf("FATAL: Failed to auto-migrate database schema: %v", err)
	}

	log.Println("Database initialized and migrated successfully.")
}

// CreateRecording inserts a new recording entry into the database.
// FIX: Renamed from CreateRecording to AddRecording and added a Filename parameter
// to align with gstreamer.go's expected call (db.AddRecording).
func AddRecording(filename string, startTime time.Time) (*Recording, error) {
	rec := &Recording{
		Filename:  filename,
		StartTime: startTime,
		Notes:     "",
		Genre:     "",
	}
	result := db.Create(rec)
	if result.Error != nil {
		log.Printf("ERROR: Failed to create recording %s: %v", filename, result.Error)
		return nil, result.Error
	}
	return rec, nil
}

// UpdateRecording updates the Notes and Genre for a specific recording ID.
func UpdateRecording(id uint, notes string, genre string) error {
	result := db.Model(&Recording{}).Where("id = ?", id).Updates(map[string]interface{}{
		"notes": notes,
		"genre": genre,
	})
	if result.Error != nil {
		log.Printf("ERROR: Failed to update recording ID %d: %v", id, result.Error)
		return result.Error
	}
	return nil
}

// CompleteRecording updates the size and duration of a recording after it finishes.
func CompleteRecording(filename string, duration time.Duration, sizeMB float64) error {
	result := db.Model(&Recording{}).Where("filename = ?", filename).Updates(map[string]interface{}{
		"duration": duration,
		"sizemb":   sizeMB,
	})
	if result.Error != nil {
		log.Printf("ERROR: Failed to complete recording %s: %v", filename, result.Error)
		return result.Error
	}
	return nil
}

// ListRecordings retrieves all recordings, sorted by start time.
func ListRecordings() ([]Recording, error) {
	var recordings []Recording
	result := db.Order("start_time desc").Find(&recordings)
	if result.Error != nil {
		log.Printf("ERROR: Failed to retrieve recordings: %v", result.Error)
		return nil, result.Error
	}
	return recordings, nil
}

// DeleteRecording deletes a recording entry by its ID.
func DeleteRecording(id uint) error {
	result := db.Delete(&Recording{}, id)
	if result.Error != nil {
		log.Printf("ERROR: Failed to delete recording ID %d: %v", id, result.Error)
		return result.Error
	}
	return nil
}

// GetRecordingByFilename retrieves a recording by its filename.
func GetRecordingByFilename(filename string) (*Recording, error) {
	var rec Recording
	result := db.Where("filename = ?", filename).First(&rec)
	if result.Error != nil {
		// Do not log error if it's a simple 'record not found'
		return nil, result.Error
	}
	return &rec, nil
}
