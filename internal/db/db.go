package db

import (
	"fmt" // ADDED: Required for error formatting
	"time"
    "context"
	"nixon/internal/slogger"
	"log/slog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"nixon/internal/common" // IMPORT: Use canonical structs
)

var dbConn *gorm.DB

// GormSlogger is a custom GORM logger that uses slog.
type GormSlogger struct {
	logger *slog.Logger
}

// NewGormSlogger creates a new GORM logger instance.
func NewGormSlogger() *GormSlogger {
	return &GormSlogger{
		// Add a "component" attribute to all logs from this logger
		logger: slogger.Log.With("component", "gorm"),
	}
}

// LogMode returns a new logger with the specified log level.
func (l *GormSlogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// We can add level switching logic here later if needed.
	return l
}

// Info logs an info message.
func (l *GormSlogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
}

// Warn logs a warning message.
func (l *GormSlogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
}

// Error logs an error message.
func (l *GormSlogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
}

// Trace logs a SQL query.
func (l *GormSlogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// Don't log successful traces if the global log level is higher than Debug.
	if err == nil && !l.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := []slog.Attr{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}

	if err != nil {
		// Add the error to the attributes and log at the Error level.
		attrs = append(attrs, slog.Any("err", err))
		l.logger.LogAttrs(ctx, slog.LevelError, "GORM query failed", attrs...)
	} else {
		l.logger.LogAttrs(ctx, slog.LevelDebug, "GORM query", attrs...)
	}
}

// Init initializes the database connection.
func Init(dsn string) error {
	var err error
	dbConn, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: NewGormSlogger().LogMode(gormlogger.Info),
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

