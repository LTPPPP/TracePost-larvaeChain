package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

func ParseQRCode(qrCodeContent string) (string, error) {
	if _, err := strconv.Atoi(qrCodeContent); err == nil {
		return qrCodeContent, nil
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(qrCodeContent), &jsonData); err == nil {
		if batchID, ok := jsonData["batch_id"].(string); ok {
			return batchID, nil
		}
		if batchID, ok := jsonData["batchId"].(string); ok {
			return batchID, nil
		}
		if batchID, ok := jsonData["id"].(string); ok {
			return batchID, nil
		}

		if batchID, ok := jsonData["batch_id"].(float64); ok {
			return strconv.FormatFloat(batchID, 'f', 0, 64), nil
		}
		if batchID, ok := jsonData["batchId"].(float64); ok {
			return strconv.FormatFloat(batchID, 'f', 0, 64), nil
		}
		if batchID, ok := jsonData["id"].(float64); ok {
			return strconv.FormatFloat(batchID, 'f', 0, 64), nil
		}
	}

	decodedData, err := base64.StdEncoding.DecodeString(qrCodeContent)
	if err == nil {
		return ParseQRCode(string(decodedData))
	}

	parts := strings.Split(qrCodeContent, ":")
	if len(parts) == 2 {
		if _, err := strconv.Atoi(parts[1]); err == nil {
			return parts[1], nil
		}
	}

	return "", errors.New("invalid QR code format")
}
