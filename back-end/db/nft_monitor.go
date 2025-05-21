package db

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// NFTLogLevel represents the severity level of a log message
type NFTLogLevel string

const (
	INFO     NFTLogLevel = "INFO"
	WARNING  NFTLogLevel = "WARNING"
	ERROR    NFTLogLevel = "ERROR"
	CRITICAL NFTLogLevel = "CRITICAL"
)

// NFTLogEntry represents a log entry for NFT operations
type NFTLogEntry struct {
	Timestamp time.Time    `json:"timestamp"`
	Level     NFTLogLevel  `json:"level"`
	NFTID     int          `json:"nft_id,omitempty"`
	TokenID   string       `json:"token_id,omitempty"`
	Operation string       `json:"operation"`
	Message   string       `json:"message"`
	Error     string       `json:"error,omitempty"`
	Data      interface{}  `json:"data,omitempty"`
}

// NFTLogger handles logging of NFT operations
type NFTLogger struct {
	LogFile  string
	LogLevel NFTLogLevel
}

// NewNFTLogger creates a new NFT logger
func NewNFTLogger() *NFTLogger {
	return &NFTLogger{
		LogFile:  getEnv("LOG_FILE", "logs/nft_operations.log"),
		LogLevel: NFTLogLevel(getEnv("LOG_LEVEL", "info")),
	}
}

// Log writes a log entry to the log file
func (l *NFTLogger) Log(entry NFTLogEntry) error {
	// Ensure log directory exists
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Open log file for appending
	file, err := os.OpenFile(l.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Set timestamp
	entry.Timestamp = time.Now()

	// Convert to JSON
	jsonEntry, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Write to file
	if _, err := file.WriteString(string(jsonEntry) + "\n"); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

// LogNFTOperation logs an NFT operation
func LogNFTOperation(level NFTLogLevel, nftID int, tokenID, operation, message string, err error, data interface{}) error {
	logger := NewNFTLogger()

	entry := NFTLogEntry{
		Level:     level,
		NFTID:     nftID,
		TokenID:   tokenID,
		Operation: operation,
		Message:   message,
		Data:      data,
	}

	if err != nil {
		entry.Error = err.Error()
	}

	return logger.Log(entry)
}

// NFTMonitor represents the monitoring system for NFT operations
type NFTMonitor struct {
	AlertThreshold int
	CheckInterval  time.Duration
}

// NewNFTMonitor creates a new NFT monitor
func NewNFTMonitor() *NFTMonitor {
	threshold := getEnvAsInt("ALERT_THRESHOLD", 5)
	interval := time.Duration(getEnvAsInt("CHECK_INTERVAL", 60)) * time.Second

	return &NFTMonitor{
		AlertThreshold: threshold,
		CheckInterval:  interval,
	}
}

// StartMonitoring begins monitoring NFT operations
func (m *NFTMonitor) StartMonitoring() {
	go func() {
		for {
			// Check for data integrity issues
			if err := m.checkDataIntegrity(); err != nil {
				LogNFTOperation(ERROR, 0, "", "monitor_integrity", "Failed to check data integrity", err, nil)
			}

			// Check for duplicate NFTs
			if err := m.checkDuplicates(); err != nil {
				LogNFTOperation(ERROR, 0, "", "monitor_duplicates", "Failed to check for duplicates", err, nil)
			}

			// Sleep for the check interval
			time.Sleep(m.CheckInterval)
		}
	}()
}

// checkDataIntegrity verifies the data integrity of NFTs
func (m *NFTMonitor) checkDataIntegrity() error {
	// Get active NFTs
	rows, err := DB.Query(`
		SELECT id, token_id FROM transaction_nft 
		WHERE is_active = true
	`)
	if err != nil {
		return fmt.Errorf("failed to query NFTs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var tokenID string
		if err := rows.Scan(&id, &tokenID); err != nil {
			return fmt.Errorf("error scanning NFT row: %w", err)
		}

		// Verify data integrity
		valid, message, err := VerifyNFTDataIntegrity(id)
		if err != nil {
			LogNFTOperation(ERROR, id, tokenID, "integrity_check", "Error verifying data integrity", err, nil)
			continue
		}

		if !valid {
			LogNFTOperation(WARNING, id, tokenID, "integrity_check", message, nil, nil)
		}
	}

	return nil
}

// checkDuplicates checks for duplicate NFTs
func (m *NFTMonitor) checkDuplicates() error {
	duplicates, err := FindDuplicateNFTs()
	if err != nil {
		return fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if len(duplicates) > 0 {
		LogNFTOperation(WARNING, 0, "", "duplicate_check", fmt.Sprintf("Found %d duplicate NFTs", len(duplicates)), nil, duplicates)
	}

	return nil
}

// AlertOnIssue sends an alert when an issue is detected
func (m *NFTMonitor) AlertOnIssue(issue string, data interface{}) error {
	LogNFTOperation(WARNING, 0, "", "alert", issue, nil, data)
	return nil
}
