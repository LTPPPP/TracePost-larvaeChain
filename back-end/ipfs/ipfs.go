package ipfs

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient represents a client for interacting with IPFS
type IPFSClient struct {
	Shell *shell.Shell
}

// NewIPFSClient creates a new IPFS client
func NewIPFSClient(apiURL string) *IPFSClient {
	return &IPFSClient{
		Shell: shell.NewShell(apiURL),
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