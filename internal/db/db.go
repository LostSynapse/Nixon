// internal/db/db.go
package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Recording represents the database model for a recording.
type Recording struct {
	ID        uint      `gorm:"primaryKey"`
	Filename  string    `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time // Automatically handled by GORM
	UpdatedAt time.Time // Automatically handled by GORM
	Name      string
	Notes     string // Changed from *string to string for simplicity with GORM defaults
	Genre     string // Changed from *string to string
	Protected bool      `gorm:"default:false"`
}

var (
	db   *gorm.DB // Unexported global variable for the database connection
	once sync.Once
	err  error // Store initialization error
)

const dbFile = "./studio.db"

// Init initializes the database connection and performs auto-migration.
// Uses sync.Once to ensure it only runs once.
func Init() error {
	once.Do(func() {
		// Configure GORM logger (optional, adjust level as needed)
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // Slow SQL threshold
				LogLevel:                  logger.Warn, // Log level (Silent, Error, Warn, Info)
				IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,       // Disable color
			},
		)

		log.Println("Initializing database connection...")
		db, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{
			Logger: newLogger,
		})
		if err != nil {
			err = fmt.Errorf("failed to connect database: %w", err)
			log.Printf("ERROR: %v", err) // Log the error during init
			return
		}

		log.Println("Running database auto-migration...")
		// AutoMigrate will create tables, missing columns, and missing indexes.
		// It will NOT delete unused columns or change existing column types.
		migrateErr := db.AutoMigrate(&Recording{})
		if migrateErr != nil {
			err = fmt.Errorf("failed to auto-migrate database: %w", migrateErr)
			log.Printf("ERROR: %v", err)
			// Close DB if migration fails?
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
			db = nil // Ensure db is nil on error
			return
		}
		log.Println("Database initialization complete.")
	})
	return err // Return the stored error from the first run
}

// GetDB returns the initialized database connection.
// Returns nil if initialization failed.
func GetDB() *gorm.DB {
	if db == nil {
		// Log might already have happened in Init()
		// log.Println("Warning: GetDB() called but database initialization failed or wasn't called.")
		return nil
	}
	return db
}

// Close closes the database connection.
func Close() {
	if db != nil {
		log.Println("Closing database connection...")
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		db = nil // Reset global var
	}
}

// --- CRUD Operations using GORM ---

// AddRecording adds a new recording to the database.
// Takes a pointer to allow GORM to populate the ID.
// Corrected signature: returns only error
func AddRecording(rec *Recording) error {
	dbConn := GetDB() // Use GetDB to ensure safe access
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}
	result := dbConn.Create(rec) // GORM handles SQL generation and execution
	if result.Error != nil {
		return fmt.Errorf("failed to add recording %s: %w", rec.Filename, result.Error)
	}
	if result.RowsAffected == 0 {
		// This might indicate a unique constraint violation or other issue
		return fmt.Errorf("failed to add recording %s (no rows affected, possibly duplicate filename?)", rec.Filename)
	}
	log.Printf("Recording '%s' added to DB with ID %d", rec.Filename, rec.ID)
	return nil // Return only error (nil on success)
}

// GetRecordings retrieves all recordings, ordered by creation date descending.
func GetRecordings() ([]Recording, error) {
	dbConn := GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var recordings []Recording
	result := dbConn.Order("created_at desc").Find(&recordings)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query recordings: %w", result.Error)
	}
	return recordings, nil
}

// UpdateRecording updates the editable fields of a recording.
func UpdateRecording(id uint, name, notes, genre string) error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}
	result := dbConn.Model(&Recording{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":  name,
		"notes": notes,
		"genre": genre,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update recording ID %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("recording ID %d not found for update", id)
	}
	log.Printf("Updated recording ID %d", id)
	return nil
}

// ToggleProtect toggles the protected status of a recording.
func ToggleProtect(id uint) error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}
	result := dbConn.Model(&Recording{}).Where("id = ?", id).Update("protected", gorm.Expr("NOT protected"))
	if result.Error != nil {
		return fmt.Errorf("failed to toggle protection for recording ID %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("recording ID %d not found for protection toggle", id)
	}
	log.Printf("Toggled protection for recording ID %d", id)
	return nil
}

// DeleteRecording deletes a recording record from the database.
// Does NOT delete the associated file. File deletion should happen first.
func DeleteRecording(id uint) error {
	dbConn := GetDB()
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}
	result := dbConn.Delete(&Recording{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete recording ID %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("recording ID %d not found for deletion", id)
	}
	log.Printf("Deleted recording record ID %d from DB", id)
	return nil
}

// GetRecording retrieves a single recording by ID (needed for file deletion check).
func GetRecording(id uint) (*Recording, error) {
	dbConn := GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var rec Recording
	result := dbConn.First(&rec, id) // Find record by primary key
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("recording ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get recording ID %d: %w", id, result.Error)
	}
	return &rec, nil
}

