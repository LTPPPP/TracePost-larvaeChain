package ipfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient represents a client for interacting with IPFS
type IPFSClient struct {
	Shell *shell.Shell
}

// IPFSService provides a higher-level interface to IPFS
type IPFSService struct {
	client *IPFSClient
}

// IPFSFile represents a file stored in IPFS
type IPFSFile struct {
	CID  string `json:"cid"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	URI  string `json:"uri"`
}

// IPFSMetadata represents JSON metadata stored in IPFS
type IPFSMetadata struct {
	CID  string `json:"cid"`
	JSON string `json:"json"`
	URI  string `json:"uri"`
}

// NewIPFSClient creates a new IPFS client
func NewIPFSClient(apiURL string) *IPFSClient {
	return &IPFSClient{
		Shell: shell.NewShell(apiURL),
	}
}

// NewIPFSService creates a new IPFS service
func NewIPFSService() *IPFSService {
	// Read IPFS node URL from environment variable or use default
	ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
	if ipfsNodeURL == "" {
		ipfsNodeURL = "http://ipfs:5001" // Default IPFS node URL
	}
	
	return &IPFSService{
		client: NewIPFSClient(ipfsNodeURL),
	}
}

// UploadFile uploads a file to IPFS
func (c *IPFSClient) UploadFile(file multipart.File) (string, error) {
	// Read file contents
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Upload to IPFS
	reader := bytes.NewReader(fileBytes)
	cid, err := c.Shell.Add(reader)
	if err != nil {
		return "", err
	}

	return cid, nil
}

// UploadJSON uploads JSON data to IPFS
func (c *IPFSClient) UploadJSON(data interface{}) (string, error) {
	// Convert data to JSON
	jsonReader, err := c.Shell.DagPut(data, "json", "cbor")
	if err != nil {
		return "", err
	}

	return jsonReader, nil
}

// StoreJSON stores JSON data on IPFS
func (s *IPFSService) StoreJSON(data interface{}) (*IPFSMetadata, error) {
	// Convert data to JSON string
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	// Upload to IPFS
	reader := bytes.NewReader(jsonBytes)
	cid, err := s.client.Shell.Add(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	
	// Get gateway URL from env or use default
	gatewayURL := os.Getenv("IPFS_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://ipfs:8080"
	}
	
	// Create metadata response
	metadata := &IPFSMetadata{
		CID:  cid,
		JSON: string(jsonBytes),
		URI:  fmt.Sprintf("%s/ipfs/%s", gatewayURL, cid),
	}
	
	return metadata, nil
}

// StoreFile stores a file on IPFS
func (s *IPFSService) StoreFile(fileData []byte, fileName string) (*IPFSFile, error) {
	// Upload to IPFS
	reader := bytes.NewReader(fileData)
	cid, err := s.client.Shell.Add(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS: %w", err)
	}
	
	// Get gateway URL from env or use default
	gatewayURL := os.Getenv("IPFS_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://ipfs:8080"
	}
	
	// Create file response
	file := &IPFSFile{
		CID:  cid,
		Name: fileName,
		Size: int64(len(fileData)),
		URI:  fmt.Sprintf("%s/ipfs/%s", gatewayURL, cid),
	}
	
	return file, nil
}

// GetFile gets a file from IPFS by its CID
func (c *IPFSClient) GetFile(cid string) ([]byte, error) {
	// Get the file from IPFS
	reader, err := c.Shell.Cat(cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Read the file contents
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return fileBytes, nil
}

// CreateIPFSURL creates a URL for accessing a file on IPFS
func (c *IPFSClient) CreateIPFSURL(cid string, gateway string) string {
	// If gateway is empty, use the default gateway
	if gateway == "" {
		gateway = "https://ipfs.io/ipfs/"
	}

	// Make sure the gateway ends with "/"
	if !strings.HasSuffix(gateway, "/") {
		gateway = gateway + "/"
	}

	return gateway + cid
}