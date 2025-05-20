package utils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// ParseQRCode trích xuất BatchId từ mã QR
// Hỗ trợ nhiều định dạng khác nhau của mã QR
func ParseQRCode(qrCodeContent string) (string, error) {
	// Trường hợp 1: Nếu mã QR chỉ chứa batchId là số nguyên
	if _, err := strconv.Atoi(qrCodeContent); err == nil {
		return qrCodeContent, nil
	}

	// Trường hợp 2: Nếu mã QR ở định dạng JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(qrCodeContent), &jsonData); err == nil {
		// Kiểm tra các trường có thể chứa batchId
		if batchID, ok := jsonData["batch_id"].(string); ok {
			return batchID, nil
		}
		if batchID, ok := jsonData["batchId"].(string); ok {
			return batchID, nil
		}
		if batchID, ok := jsonData["id"].(string); ok {
			return batchID, nil
		}
		
		// Kiểm tra các trường có thể là số
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

	// Trường hợp 3: Mã QR có thể được mã hóa base64
	decodedData, err := base64.StdEncoding.DecodeString(qrCodeContent)
	if err == nil {
		// Thử phân tích nội dung đã giải mã
		return ParseQRCode(string(decodedData))
	}

	// Trường hợp 4: Mã QR có thể có định dạng "prefix:batchId"
	parts := strings.Split(qrCodeContent, ":")
	if len(parts) == 2 {
		// Thử xác thực phần thứ hai là batchId
		if _, err := strconv.Atoi(parts[1]); err == nil {
			return parts[1], nil
		}
	}

	// Không thể xác định batchId từ mã QR
	return "", errors.New("không thể trích xuất BatchID từ mã QR")
}
