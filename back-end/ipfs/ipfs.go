package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient represents a client for interacting with IPFS
type IPFSClient struct {
	Shell       *shell.Shell
	apiURL      string
	connTimeout time.Duration
	maxRetries  int
}

// IPFSService provides a higher-level interface to IPFS

type IPFSService struct {
	client         *IPFSClient
	clientPool     []*IPFSClient
	poolSize       int
	poolMutex      sync.Mutex
	cacheEnabled   bool
	cacheTTL       time.Duration
	requestTimeout time.Duration
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

// NewIPFSClient creates a new IPFS client with optimized settings
func NewIPFSClient(apiURL string) *IPFSClient {
	shell := shell.NewShell(apiURL)
	shell.SetTimeout(30 * time.Second) // Set timeout to avoid hanging connections

	return &IPFSClient{
		Shell:       shell,
		apiURL:      apiURL,
		connTimeout: 30 * time.Second,
		maxRetries:  3, // Add retry capability for better resilience
	}
}

// NewIPFSService creates a new IPFS service with connection pooling
func NewIPFSService() *IPFSService {
	// Read IPFS node URL from environment variable or use default
	ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
	if ipfsNodeURL == "" {
		ipfsNodeURL = "http://ipfs:5001" // Default IPFS node URL
	}

	// Read pool size or use default
	poolSize := 5
	poolSizeStr := os.Getenv("IPFS_CONN_POOL_SIZE")
	if poolSizeStr != "" {
		if val, err := strconv.Atoi(poolSizeStr); err == nil && val > 0 {
			poolSize = val
		}
	}

	// Initialize the connection pool
	pool := make([]*IPFSClient, poolSize)
	for i := 0; i < poolSize; i++ {
		pool[i] = NewIPFSClient(ipfsNodeURL)
	}

	// Read cache TTL or use default (5 minutes)
	cacheTTL := 5 * time.Minute
	cacheTTLStr := os.Getenv("IPFS_CACHE_TTL")
	if cacheTTLStr != "" {
		if val, err := strconv.Atoi(cacheTTLStr); err == nil && val > 0 {
			cacheTTL = time.Duration(val) * time.Second
		}
	}

	// Read request timeout or use default (30 seconds)
	reqTimeout := 30 * time.Second
	reqTimeoutStr := os.Getenv("IPFS_REQUEST_TIMEOUT")
	if reqTimeoutStr != "" {
		if val, err := strconv.Atoi(reqTimeoutStr); err == nil && val > 0 {
			reqTimeout = time.Duration(val) * time.Second
		}
	}

	return &IPFSService{
		client:         NewIPFSClient(ipfsNodeURL),
		clientPool:     pool,
		poolSize:       poolSize,
		cacheEnabled:   os.Getenv("IPFS_CACHE_ENABLED") != "false",
		cacheTTL:       cacheTTL,
		requestTimeout: reqTimeout,
	}
}

// getClient gets a client from the pool
func (s *IPFSService) getClient() *IPFSClient {
	s.poolMutex.Lock()
	defer s.poolMutex.Unlock()

	if len(s.clientPool) == 0 {
		// If all clients are in use, create a new one
		return NewIPFSClient(s.client.apiURL)
	}

	// Get the last client from the pool
	client := s.clientPool[len(s.clientPool)-1]
	// Remove it from the pool
	s.clientPool = s.clientPool[:len(s.clientPool)-1]

	return client
}

// releaseClient returns a client to the pool
func (s *IPFSService) releaseClient(client *IPFSClient) {
	s.poolMutex.Lock()
	defer s.poolMutex.Unlock()

	// Only return to pool if we're under capacity
	if len(s.clientPool) < s.poolSize {
		s.clientPool = append(s.clientPool, client)
	}
}

// executeWithRetry executes an IPFS operation with retry logic
func (c *IPFSClient) executeWithRetry(operation func() error) error {
	var err error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		// Exponential backoff before retry
		if attempt < c.maxRetries-1 {
			backoffTime := time.Duration(attempt+1) * 500 * time.Millisecond
			time.Sleep(backoffTime)
		}
	}
	return fmt.Errorf("operation failed after %d attempts: %w", c.maxRetries, err)
}

// UploadFile uploads a file to IPFS
func (c *IPFSClient) UploadFile(file multipart.File) (string, error) {
	// Read file contents
	fileBytes, err := io.ReadAll(file)
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

// StoreJSON stores JSON data on IPFS with connection pooling, retry, and caching optimization
func (s *IPFSService) StoreJSON(data interface{}) (*IPFSMetadata, error) {
	// Get a client from the pool
	client := s.getClient()
	defer s.releaseClient(client)
	// Calculate hash of the data for potential cache key
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create a context with timeout for the operation
	_, cancel := context.WithTimeout(context.Background(), s.requestTimeout)
	defer cancel()

	var cid string

	// Upload to IPFS with retry capability
	err = client.executeWithRetry(func() error {
		reader := bytes.NewReader(jsonBytes)
		var uploadErr error
		cid, uploadErr = client.Shell.Add(reader)
		return uploadErr
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS after retries: %w", err)
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
		URI:  constructIPFSUri(gatewayURL, cid),
	}

	return metadata, nil
}

// StoreFile stores a file on IPFS with connection pooling and retries
func (s *IPFSService) StoreFile(fileData []byte, fileName string) (*IPFSFile, error) {
	// Get a client from the pool
	client := s.getClient()
	defer s.releaseClient(client)

	// Create a context with timeout for the operation
	_, cancel := context.WithTimeout(context.Background(), s.requestTimeout)
	defer cancel()

	var cid string
	var err error

	// Upload to IPFS with retry capability
	err = client.executeWithRetry(func() error {
		reader := bytes.NewReader(fileData)
		cid, err = client.Shell.Add(reader)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload to IPFS after retries: %w", err)
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
		URI:  constructIPFSUri(gatewayURL, cid),
	}

	return file, nil
}

// GetFile gets a file from IPFS by its CID with optimized performance
func (c *IPFSClient) GetFile(cid string) ([]byte, error) {
	var fileBytes []byte

	// Get the file from IPFS with retry capability
	err := c.executeWithRetry(func() error {
		reader, err := c.Shell.Cat(cid)
		if err != nil {
			return err
		}
		defer reader.Close()

		// Set a timeout for read operation
		readCtx, cancel := context.WithTimeout(context.Background(), c.connTimeout)
		defer cancel()

		// Create a channel to communicate the read result
		readDone := make(chan struct{})
		var readErr error

		go func() {
			fileBytes, readErr = io.ReadAll(reader)
			close(readDone)
		}()

		// Wait for either the read to complete or timeout
		select {
		case <-readDone:
			return readErr
		case <-readCtx.Done():
			return fmt.Errorf("read operation timed out after %v", c.connTimeout)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get file from IPFS: %w", err)
	}

	return fileBytes, nil
}

// GetFileService gets a file from IPFS using the service's connection pool
func (s *IPFSService) GetFile(cid string) ([]byte, error) {
	client := s.getClient()
	defer s.releaseClient(client)

	return client.GetFile(cid)
}

// CreateIPFSURL creates a URL for accessing a file on IPFS with flexible gateway configuration
func (c *IPFSClient) CreateIPFSURL(cid string, gateway string) string {
	// If gateway is empty, use the default gateway
	if gateway == "" {
		gateway = os.Getenv("IPFS_GATEWAY_URL")
		if gateway == "" {
			gateway = "https://ipfs.io/ipfs/"
		}
	}

	// Make sure the gateway ends with "/"
	if !strings.HasSuffix(gateway, "/") {
		gateway = gateway + "/"
	}

	// If the gateway already includes '/ipfs/', don't add it again
	if strings.Contains(gateway, "/ipfs/") {
		return gateway + cid
	}

	return gateway + "ipfs/" + cid
}

// constructIPFSUri creates a proper URI to an IPFS resource, avoiding duplicate /ipfs/ paths
func constructIPFSUri(gatewayURL string, cid string) string {
	// Remove trailing slash if present
	gatewayURL = strings.TrimSuffix(gatewayURL, "/")

	// If the gateway URL already ends with /ipfs, don't add it again
	if strings.HasSuffix(gatewayURL, "/ipfs") {
		return fmt.Sprintf("%s/%s", gatewayURL, cid)
	}

	// Otherwise add the /ipfs path
	return fmt.Sprintf("%s/ipfs/%s", gatewayURL, cid)
}

// UploadSwaggerDoc uploads the Swagger API documentation to IPFS
// and returns a URL that can be accessed via the IPFS WebUI
func (s *IPFSService) UploadSwaggerDoc(docContent []byte, docName string) (string, error) {
	// Use the existing pool management methods
	client := s.getClient()
	defer s.releaseClient(client)

	// Add the content to IPFS (no need to use ctx as the Shell.Add doesn't support it)
	cid, err := client.Shell.Add(bytes.NewReader(docContent))
	if err != nil {
		return "", fmt.Errorf("failed to add swagger doc to IPFS: %v", err)
	}

	// Get the IPFS gateway URL from environment variable
	gatewayURL := os.Getenv("IPFS_DOC_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = os.Getenv("IPFS_GATEWAY_URL")
	}

	// If gateway URL is still empty, use default
	if gatewayURL == "" {
		gatewayURL = "http://127.0.0.1:5001/webui"
	}

	// Construct the complete URL
	// Format needs to be http://127.0.0.1:5001/webui/#/explore/ipfs/CID
	if strings.HasSuffix(gatewayURL, "/") {
		gatewayURL = strings.TrimSuffix(gatewayURL, "/")
	}

	// Make sure URL format is correct - IPFS WebUI uses the #/explore/ipfs/CID format
	return fmt.Sprintf("%s/#/explore/ipfs/%s", gatewayURL, cid), nil
}

// PublishSwaggerAPI uploads the Swagger API spec to IPFS and returns a gateway URL
func PublishSwaggerAPI(specPath string) (string, error) {
	// Read Swagger spec file
	content, err := os.ReadFile(specPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Swagger spec file: %v", err)
	}

	// Create IPFS Service
	service := NewIPFSService()

	// Upload to IPFS
	url, err := service.UploadSwaggerDoc(content, "swagger.json")
	if err != nil {
		return "", err
	}

	return url, nil
}

// InitIPFSService initializes the IPFS service and ensures it is accessible
func InitIPFSService() error {
	// Create a new IPFS service
	service := NewIPFSService()
	
	// Log IPFS configuration
	ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
	ipfsGatewayURL := os.Getenv("IPFS_GATEWAY_URL")
	ipfsDocGatewayURL := os.Getenv("IPFS_DOC_GATEWAY_URL")
	
	fmt.Printf("[INFO] IPFS Configuration:\n")
	fmt.Printf("  - Node URL: %s\n", ipfsNodeURL)
	fmt.Printf("  - Gateway URL: %s\n", ipfsGatewayURL)
	fmt.Printf("  - Doc Gateway URL: %s\n", ipfsDocGatewayURL)

	// Test connection to IPFS node
	client := service.getClient()
	defer service.releaseClient(client)

	// Try to get the IPFS version to verify connectivity
	version, commit, err := client.Shell.Version()
	if err != nil {
		return fmt.Errorf("failed to connect to IPFS node: %v", err)
	}

	// Log successful connection
	fmt.Printf("[INFO] Connected to IPFS node. Version: %s, Commit: %s\n", version, commit)
	
	// Test IPFS web UI access
	webUIPath := ""
	if ipfsDocGatewayURL != "" {
		webUIPath = ipfsDocGatewayURL
	} else if ipfsGatewayURL != "" {
		webUIPath = ipfsGatewayURL
	} else {
		webUIPath = "http://127.0.0.1:5001/webui"
	}
	fmt.Printf("[INFO] IPFS WebUI should be accessible at: %s\n", webUIPath)
	
	return nil
}