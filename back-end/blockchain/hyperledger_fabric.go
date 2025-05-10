// hyperledger_fabric.go
package blockchain

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"time"
	
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// FabricClient represents a client for interacting with Hyperledger Fabric
type FabricClient struct {
	ConnectionConfig FabricConnectionConfig
	Gateway         *client.Gateway
	Network         *client.Network
	Contract        *client.Contract
	GRPCClient      *grpc.ClientConn
}

// FabricConnectionConfig contains configuration for connecting to Hyperledger Fabric
type FabricConnectionConfig struct {
	MspID             string
	CryptoPath        string
	CertPath          string
	KeyPath           string
	TlsCertPath       string
	PeerEndpoint      string
	GatewayPeer       string
	ChannelName       string
	ChaincodeName     string
	AsLocalhost       bool
}

// NewFabricClient creates a new Hyperledger Fabric client
func NewFabricClient(config FabricConnectionConfig) (*FabricClient, error) {
	client := &FabricClient{
		ConnectionConfig: config,
	}
	
	err := client.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Fabric network: %w", err)
	}
	
	return client, nil
}

// Connect establishes a connection to the Hyperledger Fabric network
func (fc *FabricClient) Connect() error {
	// Load client identity
	clientIdentity, err := fc.newIdentity()
	if err != nil {
		return fmt.Errorf("failed to create client identity: %w", err)
	}
	
	// Load client signing identity
	clientSigner, err := fc.newSigner()
	if err != nil {
		return fmt.Errorf("failed to create client signer: %w", err)
	}
	
	// Create gRPC connection to the gateway peer
	gRPCClient, err := fc.newGrpcConnection()
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	fc.GRPCClient = gRPCClient
	
	// Create Gateway connection
	gateway, err := client.Connect(
		clientIdentity,
		client.WithSign(clientSigner),
		client.WithClientConnection(gRPCClient),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	fc.Gateway = gateway
	
	// Get network and contract
	network := gateway.GetNetwork(fc.ConnectionConfig.ChannelName)
	fc.Network = network
	
	contract := network.GetContract(fc.ConnectionConfig.ChaincodeName)
	fc.Contract = contract
	
	return nil
}

// Close closes the connection to the Hyperledger Fabric network
func (fc *FabricClient) Close() {
	fc.Gateway.Close()
	fc.GRPCClient.Close()
}

// newIdentity creates a new client identity for interaction with the Fabric gateway
func (fc *FabricClient) newIdentity() (*identity.X509Identity, error) {
	certificatePEM, err := os.ReadFile(fc.ConnectionConfig.CertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	
	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	
	return identity.NewX509Identity(fc.ConnectionConfig.MspID, certificate)
}

// newSigner creates a new signing identity for interaction with the Fabric gateway
func (fc *FabricClient) newSigner() (identity.Sign, error) {
	privateKeyPEM, err := os.ReadFile(fc.ConnectionConfig.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}
	
	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	
	return identity.NewPrivateKeySign(privateKey)
}

// newGrpcConnection creates a new gRPC connection to the Fabric gateway
func (fc *FabricClient) newGrpcConnection() (*grpc.ClientConn, error) {
	tlsCertPath := fc.ConnectionConfig.TlsCertPath
	
	certificate, err := os.ReadFile(tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TLS certificate file: %w", err)
	}
	
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certificate) {
		return nil, fmt.Errorf("failed to append TLS certificate to pool")
	}
	
	transportCredentials := credentials.NewClientTLSFromCert(certPool, fc.ConnectionConfig.GatewayPeer)
	
	connection, err := grpc.Dial(
		fc.ConnectionConfig.PeerEndpoint,
		grpc.WithTransportCredentials(transportCredentials),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	
	return connection, nil
}

// CreateBatch creates a new batch on the Hyperledger Fabric blockchain
func (fc *FabricClient) CreateBatch(ctx context.Context, batchID, hatcheryID, species string, quantity int) (string, error) {
	// Create the batch data
	batchData := map[string]interface{}{
		"batch_id":     batchID,
		"hatchery_id":  hatcheryID,
		"species":      species,
		"quantity":     quantity,
		"status":       "created",
		"created_at":   time.Now().Format(time.RFC3339),
	}
	
	// Convert to JSON
	batchDataJSON, err := json.Marshal(batchData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal batch data: %w", err)
	}
	
	// Submit transaction to the blockchain
	result, err := fc.Contract.SubmitTransaction("CreateBatch", string(batchDataJSON))
	if err != nil {
		return "", fmt.Errorf("failed to submit CreateBatch transaction: %w", err)
	}
	
	return string(result), nil
}

// UpdateBatchStatus updates the status of a batch on the Hyperledger Fabric blockchain
func (fc *FabricClient) UpdateBatchStatus(ctx context.Context, batchID, status string) (string, error) {
	// Create the update data
	updateData := map[string]interface{}{
		"batch_id":   batchID,
		"status":     status,
		"updated_at": time.Now().Format(time.RFC3339),
	}
	
	// Convert to JSON
	updateDataJSON, err := json.Marshal(updateData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal update data: %w", err)
	}
	
	// Submit transaction to the blockchain
	result, err := fc.Contract.SubmitTransaction("UpdateBatchStatus", string(updateDataJSON))
	if err != nil {
		return "", fmt.Errorf("failed to submit UpdateBatchStatus transaction: %w", err)
	}
	
	return string(result), nil
}

// GetBatchDetails gets the details of a batch from the Hyperledger Fabric blockchain
func (fc *FabricClient) GetBatchDetails(ctx context.Context, batchID string) (map[string]interface{}, error) {
	// Evaluate transaction
	result, err := fc.Contract.EvaluateTransaction("GetBatchDetails", batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate GetBatchDetails transaction: %w", err)
	}
	
	// Parse result
	var batchDetails map[string]interface{}
	err = json.Unmarshal(result, &batchDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch details: %w", err)
	}
	
	return batchDetails, nil
}

// GetBatchHistory gets the history of a batch from the Hyperledger Fabric blockchain
func (fc *FabricClient) GetBatchHistory(ctx context.Context, batchID string) ([]map[string]interface{}, error) {
	// Evaluate transaction
	result, err := fc.Contract.EvaluateTransaction("GetBatchHistory", batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate GetBatchHistory transaction: %w", err)
	}
	
	// Parse result
	var batchHistory []map[string]interface{}
	err = json.Unmarshal(result, &batchHistory)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal batch history: %w", err)
	}
	
	return batchHistory, nil
}

// QueryBatches queries batches from the Hyperledger Fabric blockchain
func (fc *FabricClient) QueryBatches(ctx context.Context, queryString string) ([]map[string]interface{}, error) {
	// Evaluate transaction
	result, err := fc.Contract.EvaluateTransaction("QueryBatches", queryString)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate QueryBatches transaction: %w", err)
	}
	
	// Parse result
	var batches []map[string]interface{}
	err = json.Unmarshal(result, &batches)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal batches: %w", err)
	}
	
	return batches, nil
}
