package ipfs

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"sync"
)

// IPFSPinataService combines IPFS and Pinata services
type IPFSPinataService struct {
	ipfsService   *IPFSService
	pinataService *PinataService
	autoPinToPinata bool
	mu           sync.Mutex
}

// IPFSPinataResult represents the result of an IPFS+Pinata operation
type IPFSPinataResult struct {
	CID           string `json:"cid"`
	Name          string `json:"name"`
	Size          int64  `json:"size"`
	IPFSUri       string `json:"ipfsUri"`
	PinataUri     string `json:"pinataUri,omitempty"`
	PinataSuccess bool   `json:"pinataPinned"`
}

// NewIPFSPinataService creates a new combined service
func NewIPFSPinataService() *IPFSPinataService {
	// Check if Pinata auto-pinning is enabled
	autoPinToPinata := true
	if autoPin := os.Getenv("IPFS_AUTO_PIN_TO_PINATA"); autoPin == "false" {
		autoPinToPinata = false
	}
	
	return &IPFSPinataService{
		ipfsService:    NewIPFSService(),
		pinataService:  NewPinataService(),
		autoPinToPinata: autoPinToPinata,
	}
}

// GetPinataService returns the underlying PinataService for validation
func (s *IPFSPinataService) GetPinataService() *PinataService {
	return s.pinataService
}

// UploadFile uploads a file to IPFS and optionally pins it to Pinata
func (s *IPFSPinataService) UploadFile(file multipart.File, filename string, metadata map[string]string, pinToPinata bool) (*IPFSPinataResult, error) {
	// Create a copy of the file content
	contentCopy, err := copyMultipartFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %v", err)
	}
	
	// Reset the original file for IPFS upload
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to reset file position: %v", err)
	}
	
	// Get a client from the pool
	ipfsClient := s.ipfsService.getClient()
	defer s.ipfsService.releaseClient(ipfsClient)
	
	// Upload to IPFS
	cid, err := ipfsClient.UploadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to IPFS: %v", err)
	}
	
	// Create the result
	result := &IPFSPinataResult{
		CID:          cid,
		Name:         filename,
		IPFSUri:      s.ipfsService.client.CreateIPFSURL(cid, ""),
		PinataSuccess: false,
	}
	
	// Pin to Pinata if requested
	if (pinToPinata || s.autoPinToPinata) && s.pinataService != nil {
		// Use the copy for Pinata upload
		if _, err := contentCopy.Seek(0, io.SeekStart); err != nil {
			return result, fmt.Errorf("failed to reset file position for Pinata: %v", err)
		}
		
		pinResponse, err := s.pinataService.PinFile(contentCopy, filename, metadata)
		if err != nil {
			// Return partial success if IPFS upload worked but Pinata failed
			return result, fmt.Errorf("file uploaded to IPFS but failed to pin to Pinata: %v", err)
		}
		
		// Update the result with Pinata info
		result.PinataUri = s.pinataService.CreatePinataGatewayURL(pinResponse.IpfsHash)
		result.PinataSuccess = true
	}
	
	return result, nil
}

// UploadJSON uploads JSON to IPFS and optionally pins it to Pinata
func (s *IPFSPinataService) UploadJSON(data interface{}, name string, metadata map[string]string, pinToPinata bool) (*IPFSPinataResult, error) {
	// Get a client from the pool
	ipfsClient := s.ipfsService.getClient()
	defer s.ipfsService.releaseClient(ipfsClient)
	
	// Upload to IPFS
	cid, err := ipfsClient.UploadJSON(data)
	if err != nil {
		fmt.Printf("IPFS JSON upload error: %v\n", err)
		// We'll continue even with an error to try Pinata directly
	}
	
	// Create the result
	result := &IPFSPinataResult{
		CID:          cid,
		Name:         name,
		IPFSUri:      "",
		PinataSuccess: false,
	}
	
	// Set the IPFS URI if we have a CID
	if cid != "" {
		result.IPFSUri = s.ipfsService.client.CreateIPFSURL(cid, "")
	}
	
	// Always try to pin to Pinata if service is available
	if s.pinataService != nil {
		pinResponse, err := s.pinataService.PinJSON(data, name, metadata)
		if err != nil {
			fmt.Printf("Pinata JSON pin error: %v\n", err)
			// Still return the result with IPFS info
		} else {
			// Update the result with Pinata info
			result.PinataUri = s.pinataService.CreatePinataGatewayURL(pinResponse.IpfsHash)
			result.PinataSuccess = true
			
			// If IPFS upload failed but Pinata worked, update the CID
			if cid == "" && pinResponse.IpfsHash != "" {
				result.CID = pinResponse.IpfsHash
				// Update the IPFS URI now that we have a CID
				result.IPFSUri = s.ipfsService.client.CreateIPFSURL(pinResponse.IpfsHash, "")
			}
			
			fmt.Printf("Successfully pinned to Pinata with URI: %s\n", result.PinataUri)
		}
	}
	
	return result, nil
}

// PinExistingCIDToPinata pins an existing IPFS CID to Pinata
func (s *IPFSPinataService) PinExistingCIDToPinata(cid string, name string, metadata map[string]string) (*IPFSPinataResult, error) {
	// Make sure Pinata service is initialized
	if s.pinataService == nil {
		return nil, fmt.Errorf("Pinata service not initialized")
	}
	
	// Pin to Pinata
	pinResponse, err := s.pinataService.PinByCID(cid, name, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to pin CID to Pinata: %v", err)
	}
	
	// Create the result
	result := &IPFSPinataResult{
		CID:          cid,
		Name:         name,
		IPFSUri:      s.ipfsService.client.CreateIPFSURL(cid, ""),
		PinataUri:    s.pinataService.CreatePinataGatewayURL(pinResponse.IpfsHash),
		PinataSuccess: true,
	}
	
	return result, nil
}

// UploadSwaggerDocWithPinata uploads the Swagger API documentation to IPFS and Pinata
func (s *IPFSPinataService) UploadSwaggerDocWithPinata(docContent []byte, docName string) (string, error) {
	// Use the existing IPFS service to upload the document
	ipfsUrl, err := s.ipfsService.UploadSwaggerDoc(docContent, docName)
	if err != nil {
		return "", fmt.Errorf("failed to upload swagger doc to IPFS: %v", err)
	}
	
	// Extract the CID from the IPFS URL
	// Format is http://127.0.0.1:5001/webui/#/explore/ipfs/CID
	parts := extractCIDFromWebUIUrl(ipfsUrl)
	if parts == "" {
		return ipfsUrl, fmt.Errorf("could not extract CID from IPFS URL, but upload succeeded")
	}
	
	// Pin to Pinata if auto-pinning is enabled
	if s.autoPinToPinata && s.pinataService != nil {
		metadata := map[string]string{
			"type": "swagger-doc",
			"name": docName,
		}
		
		_, err := s.pinataService.PinByCID(parts, "Swagger API Documentation", metadata)
		if err != nil {
			// Return the original URL but with a warning
			return ipfsUrl, fmt.Errorf("document uploaded to IPFS but failed to pin to Pinata: %v", err)
		}
		
		// Return the Pinata gateway URL instead
		pinataUrl := s.pinataService.CreatePinataGatewayURL(parts)
		return pinataUrl, nil
	}
	
	return ipfsUrl, nil
}

// InitIPFSPinataService initializes the IPFS and Pinata services
func InitIPFSPinataService() error {
	// Initialize IPFS service
	if err := InitIPFSService(); err != nil {
		return err
	}
	
	// Create a Pinata service
	pinataService := NewPinataService()
	
	// Check if the Pinata integration is properly configured
	if pinataService.JWT != "" || (pinataService.APIKey != "" && pinataService.APISecret != "") {
		// Test Pinata connection
		if err := pinataService.TestPinataConnection(); err != nil {
			fmt.Printf("[WARN] Pinata connection test failed: %v\n", err)
			fmt.Println("[WARN] Pinata integration will be disabled")
		} else {
			fmt.Println("[INFO] Pinata integration is active and working")
			fmt.Printf("[INFO] Pinata Gateway URL: %s\n", pinataService.GatewayURL)
		}
	} else {
		fmt.Println("[INFO] Pinata integration not configured (JWT or API Key/Secret missing)")
	}
	
	return nil
}

// extractCIDFromWebUIUrl extracts the CID from an IPFS WebUI URL
func extractCIDFromWebUIUrl(url string) string {
	// Extract the CID from URLs like http://127.0.0.1:5001/webui/#/explore/ipfs/CID
	parts := splitLast(url, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

// splitLast splits a string by the last occurrence of a separator
func splitLast(s, sep string) []string {
	i := len(s) - 1
	for i >= 0 && s[i:i+1] != sep {
		i--
	}
	if i < 0 {
		return []string{s}
	}
	return []string{s[:i], s[i+1:]}
}

// copyMultipartFile creates a copy of a multipart file
func copyMultipartFile(src multipart.File) (multipart.File, error) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "ipfs-upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	
	// Copy the content to the temporary file
	if _, err := io.Copy(tmpFile, src); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to copy content to temporary file: %v", err)
	}
	
	// Reset the file position
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to reset file position: %v", err)
	}
	
	// The temporary file will be automatically removed when the program exits
	// or when the file is closed
	
	return tmpFile, nil
}
