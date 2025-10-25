package db

import (
	"fmt" // ADDED: Required for error formatting
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"nixon/internal/common" // IMPORT: Use canonical structs
)

var dbConn *gorm.DB

// Init initializes the database connection.
func Init(dsn string) error {
	var err error
	dbConn, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	// Auto-migrate the database schema using the canonical struct
	return dbConn.AutoMigrate(&common.Recording{})
}

// AddRecording creates a new recording entry in the database.
// It now returns the common.Recording struct.
func AddRecording(filename string, startTime time.Time) (*common.Recording, error) {
	rec := &common.Recording{
		Filename:  filename,
		StartTime: startTime,
		Notes:     "", // Default value
		Genre:     "", // Default value
	}
	result := dbConn.Create(rec)
	if result.Error != nil {
		return nil, result.Error
	}
	return rec, nil
}

// UpdateRecording updates an existing recording in the database.
func UpdateRecording(id uint, notes, genre string, endTime time.Time, duration time.Duration) error {
	if dbConn == nil {
		return fmt.Errorf("database not initialized")
	}

	// Find the recording first
	var rec common.Recording
	if err := dbConn.First(&rec, id).Error; err != nil {
		return err // Handle not found error
	}

	// Update fields
	rec.Notes = notes
	rec.Genre = genre
	rec.EndTime = endTime
	rec.Duration = duration

	// Save the updated record
	return dbConn.Save(&rec).Error
}

// DeleteRecording removes a recording from the database.
func DeleteRecording(id uint) error {
	result := dbConn.Delete(&common.Recording{}, id)
	return result.Error
}

// GetAllRecordings retrieves all recording entries.
func GetAllRecordings() ([]common.Recording, error) {
	var recordings []common.Recording
	result := dbConn.Find(&recordings)
	return recordings, result.Error
}

// GetRecordingByID retrieves a single recording by its ID.
func GetRecordingByID(id uint) (*common.Recording, error) {
	var rec common.Recording
	result := dbConn.First(&rec, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &rec, nil
}

