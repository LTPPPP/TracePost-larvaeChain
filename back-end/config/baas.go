// Package config provides configuration for the BaaS service
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// BaaSConfig represents the configuration for Blockchain-as-a-Service
type BaaSConfig struct {
	// General configuration
	ServiceID      string `json:"service_id"`
	ServiceName    string `json:"service_name"`
	ServiceVersion string `json:"service_version"`
	LogLevel       string `json:"log_level"`
	APIKey         string `json:"api_key,omitempty"`
	APIEndpoint    string `json:"api_endpoint"`
	
	// Network configurations
	Networks []NetworkConfig `json:"networks"`
	
	// Cross-chain configurations
	CrossChainConfig CrossChainConfig `json:"cross_chain_config"` 
	
	// Deployment configuration
	DeploymentConfig DeploymentConfig `json:"deployment_config"`
	
	// Security configuration
	SecurityConfig SecurityConfig `json:"security_config"`
	
	// Monitoring configuration
	MonitoringConfig MonitoringConfig `json:"monitoring_config"`
	
	// IPFS configuration
	IPFSConfig IPFSConfig `json:"ipfs_config"`
	
	// Infrastructure configuration
	InfrastructureConfig InfrastructureConfig `json:"infrastructure_config"`
	
	// API configuration
	APIConfig APIConfig `json:"api_config"`
	
	// Fallback services
	FallbackServices []FallbackService `json:"fallback_services"`
	
	// Governance configuration
	GovernanceConfig GovernanceConfig `json:"governance_config"`
	
	// Lock for thread-safe operations
	mutex sync.RWMutex
}

// NetworkConfig represents the configuration for a blockchain network
type NetworkConfig struct {
	NetworkID          string                 `json:"network_id"`
	NetworkName        string                 `json:"network_name"`
	NetworkType        string                 `json:"network_type"` // "ethereum", "polkadot", "cosmos", "hyperledger", etc.
	NetworkVersion     string                 `json:"network_version"`
	Endpoints          []string               `json:"endpoints"`
	ChainID            string                 `json:"chain_id"`
	BootNodes          []string               `json:"boot_nodes,omitempty"`
	ConsensusAlgorithm string                 `json:"consensus_algorithm"`
	NetworkParams      map[string]interface{} `json:"network_params"`
	ContractAddresses  map[string]string      `json:"contract_addresses"`
	ApiKeys            map[string]string      `json:"api_keys,omitempty"`
	GasPrice           string                 `json:"gas_price,omitempty"`
	GasLimit           uint64                 `json:"gas_limit,omitempty"`
	Enabled            bool                   `json:"enabled"`
}

// CrossChainConfig represents the configuration for cross-chain interactions
type CrossChainConfig struct {
	Enabled                bool                   `json:"enabled"`
	BridgeConfigurations   []BridgeConfiguration  `json:"bridge_configurations"`
	InteroperabilityConfig map[string]interface{} `json:"interoperability_config"`
	SupportedProtocols     []string               `json:"supported_protocols"` // "IBC", "XCM", "Hash Time Lock", etc.
	AssetMappings          []AssetMapping         `json:"asset_mappings"`
	TrustedRelayers        []string               `json:"trusted_relayers"`
	OracleURLs             []string               `json:"oracle_urls"`
	ValidationThreshold    int                    `json:"validation_threshold"`
	MaxPacketSize          int                    `json:"max_packet_size"`
	MaxTimeoutHeight       uint64                 `json:"max_timeout_height"`
	RouteCache             map[string]interface{} `json:"route_cache"`
}

// BridgeConfiguration represents the configuration for a specific bridge
type BridgeConfiguration struct {
	BridgeID           string                 `json:"bridge_id"`
	BridgeName         string                 `json:"bridge_name"`
	BridgeType         string                 `json:"bridge_type"` // "ibc", "xcm", "hash_time_lock", etc.
	SourceNetworkID    string                 `json:"source_network_id"`
	DestinationNetworkID string               `json:"destination_network_id"`
	ContractAddresses  map[string]string      `json:"contract_addresses"`
	EndpointURLs       []string               `json:"endpoint_urls"`
	AdditionalParams   map[string]interface{} `json:"additional_params"`
	Enabled            bool                   `json:"enabled"`
	Fee                string                 `json:"fee,omitempty"`
	GasLimit           uint64                 `json:"gas_limit,omitempty"`
	VerificationMode   string                 `json:"verification_mode"` // "optimistic", "zk", "validity"
}

// AssetMapping represents the mapping of assets across chains
type AssetMapping struct {
	AssetID        string                 `json:"asset_id"`
	AssetName      string                 `json:"asset_name"`
	AssetSymbol    string                 `json:"asset_symbol"`
	Decimals       int                    `json:"decimals"`
	ChainMappings  []ChainAssetMapping    `json:"chain_mappings"`
	Metadata       map[string]interface{} `json:"metadata"`
	VerificationURI string                `json:"verification_uri,omitempty"`
}

// ChainAssetMapping represents a mapping of an asset on a specific chain
type ChainAssetMapping struct {
	ChainID           string            `json:"chain_id"`
	LocalAssetID      string            `json:"local_asset_id"`
	ContractAddress   string            `json:"contract_address,omitempty"`
	MultilocationPath map[string]interface{} `json:"multilocation_path,omitempty"`
	AssetType         string            `json:"asset_type"` // "native", "erc20", "erc721", "substrate", etc.
	IsNative          bool              `json:"is_native"`
	BridgeID          string            `json:"bridge_id,omitempty"`
}

// DeploymentConfig represents the configuration for deployment
type DeploymentConfig struct {
	Environment       string            `json:"environment"` // "development", "staging", "production"
	MinimumNodes      int               `json:"minimum_nodes"`
	RecommendedNodes  int               `json:"recommended_nodes"`
	HighAvailability  bool              `json:"high_availability"`
	ResourceLimits    map[string]string `json:"resource_limits"`
	AutoScaling       bool              `json:"auto_scaling"`
	DeploymentRegions []string          `json:"deployment_regions"`
	K8sNamespace      string            `json:"k8s_namespace,omitempty"`
	DockerRegistry    string            `json:"docker_registry,omitempty"`
	StorageClass      string            `json:"storage_class,omitempty"`
}

// SecurityConfig represents the security configuration
type SecurityConfig struct {
	RateLimiting         bool              `json:"rate_limiting"`
	MaxRequestsPerMinute int               `json:"max_requests_per_minute"`
	IPWhitelist          []string          `json:"ip_whitelist"`
	IPBlacklist          []string          `json:"ip_blacklist"`
	AuthMechanisms       []string          `json:"auth_mechanisms"` // "jwt", "oauth", "apikey", etc.
	JWTIssuer            string            `json:"jwt_issuer,omitempty"`
	OAuthProviders       []string          `json:"oauth_providers,omitempty"`
	CORSOrigins          []string          `json:"cors_origins"`
	AllowedActions       map[string]string `json:"allowed_actions"` // role-based actions
	EncryptionKeys       map[string]string `json:"encryption_keys,omitempty"`
	AuditLogging         bool              `json:"audit_logging"`
	TLSEnabled           bool              `json:"tls_enabled"`
	TLSCertPath          string            `json:"tls_cert_path,omitempty"`
	TLSKeyPath           string            `json:"tls_key_path,omitempty"`
}

// MonitoringConfig represents the monitoring configuration
type MonitoringConfig struct {
	Enabled               bool     `json:"enabled"`
	PrometheusEndpoint    string   `json:"prometheus_endpoint,omitempty"`
	AlertManagerEndpoint  string   `json:"alert_manager_endpoint,omitempty"`
	GrafanaDashboardURL   string   `json:"grafana_dashboard_url,omitempty"`
	NotificationEmails    []string `json:"notification_emails"`
	SlackWebhook          string   `json:"slack_webhook,omitempty"`
	PagerDutyIntegration  string   `json:"pager_duty_integration,omitempty"`
	LoggingLevel          string   `json:"logging_level"`
	MetricsEnabled        bool     `json:"metrics_enabled"`
	TracingEnabled        bool     `json:"tracing_enabled"`
	JaegerEndpoint        string   `json:"jaeger_endpoint,omitempty"`
	SentryDSN             string   `json:"sentry_dsn,omitempty"`
	HealthCheckInterval   int      `json:"health_check_interval"`
	PerformanceThresholds map[string]interface{} `json:"performance_thresholds"`
}

// IPFSConfig represents the IPFS configuration
type IPFSConfig struct {
	Enabled           bool     `json:"enabled"`
	APIEndpoint       string   `json:"api_endpoint"`
	GatewayEndpoint   string   `json:"gateway_endpoint"`
	PinningServices   []string `json:"pinning_services,omitempty"`
	ReplicationFactor int      `json:"replication_factor"`
	MaxFileSize       int64    `json:"max_file_size"`
	AllowedMimeTypes  []string `json:"allowed_mime_types"`
	EncryptionEnabled bool     `json:"encryption_enabled"`
	AuthType          string   `json:"auth_type,omitempty"` // "none", "basic", "jwt", etc.
	Username          string   `json:"username,omitempty"`
	Password          string   `json:"password,omitempty"`
	JWTToken          string   `json:"jwt_token,omitempty"`
}

// InfrastructureConfig represents the infrastructure configuration
type InfrastructureConfig struct {
	CloudProvider          string            `json:"cloud_provider,omitempty"` // "aws", "gcp", "azure", etc.
	Region                 string            `json:"region,omitempty"`
	VPC                    string            `json:"vpc,omitempty"`
	Subnet                 string            `json:"subnet,omitempty"`
	ContainerRepository    string            `json:"container_repository,omitempty"`
	KubernetesCluster      string            `json:"kubernetes_cluster,omitempty"`
	NodeSelector           map[string]string `json:"node_selector,omitempty"`
	ResourceRequirements   map[string]string `json:"resource_requirements"`
	StorageClass           string            `json:"storage_class,omitempty"`
	StorageSize            string            `json:"storage_size"`
	BackupEnabled          bool              `json:"backup_enabled"`
	BackupSchedule         string            `json:"backup_schedule,omitempty"`
	BackupRetentionDays    int               `json:"backup_retention_days"`
	DisasterRecoveryEnabled bool             `json:"disaster_recovery_enabled"`
	HAEnabled              bool              `json:"ha_enabled"`
}

// APIConfig represents the API configuration
type APIConfig struct {
	RestEnabled            bool     `json:"rest_enabled"`
	RestPort               int      `json:"rest_port"`
	GraphQLEnabled         bool     `json:"graphql_enabled"`
	GraphQLPort            int      `json:"graphql_port"`
	WebsocketEnabled       bool     `json:"websocket_enabled"`
	WebsocketPort          int      `json:"websocket_port"`
	RateLimit              int      `json:"rate_limit"`
	DocEnabled             bool     `json:"doc_enabled"`
	DocPath                string   `json:"doc_path,omitempty"`
	VersionPath            string   `json:"version_path,omitempty"`
	HealthCheckPath        string   `json:"health_check_path"`
	APIKeys                []string `json:"api_keys,omitempty"`
	RequestTimeout         int      `json:"request_timeout"`
	MaxRequestBodySize     int      `json:"max_request_body_size"`
	EnableCompression      bool     `json:"enable_compression"`
	AllowedOrigins         []string `json:"allowed_origins"`
	AllowedMethods         []string `json:"allowed_methods"`
	AllowedHeaders         []string `json:"allowed_headers"`
	ExposedHeaders         []string `json:"exposed_headers"`
	AllowCredentials       bool     `json:"allow_credentials"`
}

// FallbackService represents a fallback service
type FallbackService struct {
	ServiceID      string   `json:"service_id"`
	ServiceName    string   `json:"service_name"`
	ServiceType    string   `json:"service_type"`
	Endpoints      []string `json:"endpoints"`
	Priority       int      `json:"priority"`
	HealthCheckURL string   `json:"health_check_url"`
	APIKeyHeader   string   `json:"api_key_header,omitempty"`
	APIKey         string   `json:"api_key,omitempty"`
	Timeout        int      `json:"timeout"`
	Enabled        bool     `json:"enabled"`
}

// GovernanceConfig represents the governance configuration
type GovernanceConfig struct {
	VotingEnabled       bool     `json:"voting_enabled"`
	ProposalThreshold   int      `json:"proposal_threshold"`
	VotingPeriod        int      `json:"voting_period"` // in blocks or time units
	VotingQuorum        int      `json:"voting_quorum"` // percentage
	EnactmentDelay      int      `json:"enactment_delay"`
	GovernanceContract  string   `json:"governance_contract,omitempty"`
	GovernanceMembers   []string `json:"governance_members,omitempty"`
	EmergencyCommittee  []string `json:"emergency_committee,omitempty"`
	ProposalTypes       []string `json:"proposal_types"` // "parameter", "upgrade", "asset", etc.
	VetoThreshold       int      `json:"veto_threshold"`
	EmergencyVetoEnabled bool    `json:"emergency_veto_enabled"`
}

// LoadBaaSConfig loads the BaaS configuration from a file
func LoadBaaSConfig(configPath string) (*BaaSConfig, error) {
	// Check if the file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}
	
	// Read the config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}
	
	// Parse the config
	var config BaaSConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}
	
	return &config, nil
}

// SaveBaaSConfig saves the BaaS configuration to a file
func (c *BaaSConfig) SaveBaaSConfig(configPath string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}
	
	// Marshal the config to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	
	// Write the config to the file
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	
	return nil
}

// GetNetworkConfig returns the configuration for a specific network
func (c *BaaSConfig) GetNetworkConfig(networkID string) (*NetworkConfig, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, network := range c.Networks {
		if network.NetworkID == networkID {
			return &network, nil
		}
	}
	
	return nil, fmt.Errorf("network not found: %s", networkID)
}

// AddNetworkConfig adds a new network configuration
func (c *BaaSConfig) AddNetworkConfig(network NetworkConfig) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Check if the network already exists
	for i, n := range c.Networks {
		if n.NetworkID == network.NetworkID {
			// Update the existing network
			c.Networks[i] = network
			return
		}
	}
	
	// Add the new network
	c.Networks = append(c.Networks, network)
}

// RemoveNetworkConfig removes a network configuration
func (c *BaaSConfig) RemoveNetworkConfig(networkID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, network := range c.Networks {
		if network.NetworkID == networkID {
			// Remove the network
			c.Networks = append(c.Networks[:i], c.Networks[i+1:]...)
			return
		}
	}
}

// GetBridgeConfiguration returns the configuration for a specific bridge
func (c *BaaSConfig) GetBridgeConfiguration(bridgeID string) (*BridgeConfiguration, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, bridge := range c.CrossChainConfig.BridgeConfigurations {
		if bridge.BridgeID == bridgeID {
			return &bridge, nil
		}
	}
	
	return nil, fmt.Errorf("bridge not found: %s", bridgeID)
}

// AddBridgeConfiguration adds a new bridge configuration
func (c *BaaSConfig) AddBridgeConfiguration(bridge BridgeConfiguration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Check if the bridge already exists
	for i, b := range c.CrossChainConfig.BridgeConfigurations {
		if b.BridgeID == bridge.BridgeID {
			// Update the existing bridge
			c.CrossChainConfig.BridgeConfigurations[i] = bridge
			return
		}
	}
	
	// Add the new bridge
	c.CrossChainConfig.BridgeConfigurations = append(c.CrossChainConfig.BridgeConfigurations, bridge)
}

// RemoveBridgeConfiguration removes a bridge configuration
func (c *BaaSConfig) RemoveBridgeConfiguration(bridgeID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, bridge := range c.CrossChainConfig.BridgeConfigurations {
		if bridge.BridgeID == bridgeID {
			// Remove the bridge
			c.CrossChainConfig.BridgeConfigurations = append(c.CrossChainConfig.BridgeConfigurations[:i], c.CrossChainConfig.BridgeConfigurations[i+1:]...)
			return
		}
	}
}

// GetAssetMapping returns the mapping for a specific asset
func (c *BaaSConfig) GetAssetMapping(assetID string) (*AssetMapping, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, mapping := range c.CrossChainConfig.AssetMappings {
		if mapping.AssetID == assetID {
			return &mapping, nil
		}
	}
	
	return nil, fmt.Errorf("asset mapping not found: %s", assetID)
}

// AddAssetMapping adds a new asset mapping
func (c *BaaSConfig) AddAssetMapping(mapping AssetMapping) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Check if the mapping already exists
	for i, m := range c.CrossChainConfig.AssetMappings {
		if m.AssetID == mapping.AssetID {
			// Update the existing mapping
			c.CrossChainConfig.AssetMappings[i] = mapping
			return
		}
	}
	
	// Add the new mapping
	c.CrossChainConfig.AssetMappings = append(c.CrossChainConfig.AssetMappings, mapping)
}

// RemoveAssetMapping removes an asset mapping
func (c *BaaSConfig) RemoveAssetMapping(assetID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, mapping := range c.CrossChainConfig.AssetMappings {
		if mapping.AssetID == assetID {
			// Remove the mapping
			c.CrossChainConfig.AssetMappings = append(c.CrossChainConfig.AssetMappings[:i], c.CrossChainConfig.AssetMappings[i+1:]...)
			return
		}
	}
}

// GetChainAssetMapping returns the chain-specific mapping for an asset
func (c *BaaSConfig) GetChainAssetMapping(assetID, chainID string) (*ChainAssetMapping, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, mapping := range c.CrossChainConfig.AssetMappings {
		if mapping.AssetID == assetID {
			for _, chainMapping := range mapping.ChainMappings {
				if chainMapping.ChainID == chainID {
					return &chainMapping, nil
				}
			}
			return nil, fmt.Errorf("chain mapping not found for asset %s on chain %s", assetID, chainID)
		}
	}
	
	return nil, fmt.Errorf("asset mapping not found: %s", assetID)
}

// AddChainAssetMapping adds a chain-specific mapping for an asset
func (c *BaaSConfig) AddChainAssetMapping(assetID string, chainMapping ChainAssetMapping) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, mapping := range c.CrossChainConfig.AssetMappings {
		if mapping.AssetID == assetID {
			// Check if the chain mapping already exists
			for j, cm := range mapping.ChainMappings {
				if cm.ChainID == chainMapping.ChainID {
					// Update the existing chain mapping
					c.CrossChainConfig.AssetMappings[i].ChainMappings[j] = chainMapping
					return nil
				}
			}
			
			// Add the new chain mapping
			c.CrossChainConfig.AssetMappings[i].ChainMappings = append(c.CrossChainConfig.AssetMappings[i].ChainMappings, chainMapping)
			return nil
		}
	}
	
	return fmt.Errorf("asset mapping not found: %s", assetID)
}

// RemoveChainAssetMapping removes a chain-specific mapping for an asset
func (c *BaaSConfig) RemoveChainAssetMapping(assetID, chainID string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, mapping := range c.CrossChainConfig.AssetMappings {
		if mapping.AssetID == assetID {
			for j, cm := range mapping.ChainMappings {
				if cm.ChainID == chainID {
					// Remove the chain mapping
					c.CrossChainConfig.AssetMappings[i].ChainMappings = append(
						c.CrossChainConfig.AssetMappings[i].ChainMappings[:j],
						c.CrossChainConfig.AssetMappings[i].ChainMappings[j+1:]...)
					return nil
				}
			}
			return fmt.Errorf("chain mapping not found for asset %s on chain %s", assetID, chainID)
		}
	}
	
	return fmt.Errorf("asset mapping not found: %s", assetID)
}

// EnableCrossChain enables cross-chain functionality
func (c *BaaSConfig) EnableCrossChain(enabled bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.CrossChainConfig.Enabled = enabled
}

// IsCrossChainEnabled returns whether cross-chain functionality is enabled
func (c *BaaSConfig) IsCrossChainEnabled() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	return c.CrossChainConfig.Enabled
}

// GetSupportedNetworks returns a list of enabled network IDs
func (c *BaaSConfig) GetSupportedNetworks() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	networks := []string{}
	for _, network := range c.Networks {
		if network.Enabled {
			networks = append(networks, network.NetworkID)
		}
	}
	
	return networks
}

// GetSupportedBridges returns a list of enabled bridge IDs
func (c *BaaSConfig) GetSupportedBridges() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	bridges := []string{}
	for _, bridge := range c.CrossChainConfig.BridgeConfigurations {
		if bridge.Enabled {
			bridges = append(bridges, bridge.BridgeID)
		}
	}
	
	return bridges
}

// GetNetworkByType returns the first network of a specific type
func (c *BaaSConfig) GetNetworkByType(networkType string) (*NetworkConfig, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, network := range c.Networks {
		if strings.ToLower(network.NetworkType) == strings.ToLower(networkType) && network.Enabled {
			return &network, nil
		}
	}
	
	return nil, fmt.Errorf("no enabled network of type %s found", networkType)
}

// GetBridgesByType returns all bridges of a specific type
func (c *BaaSConfig) GetBridgesByType(bridgeType string) []BridgeConfiguration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	bridges := []BridgeConfiguration{}
	for _, bridge := range c.CrossChainConfig.BridgeConfigurations {
		if strings.ToLower(bridge.BridgeType) == strings.ToLower(bridgeType) && bridge.Enabled {
			bridges = append(bridges, bridge)
		}
	}
	
	return bridges
}

// GetContractAddress returns the contract address for a specific network and contract name
func (c *BaaSConfig) GetContractAddress(networkID, contractName string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, network := range c.Networks {
		if network.NetworkID == networkID {
			address, ok := network.ContractAddresses[contractName]
			if !ok {
				return "", fmt.Errorf("contract %s not found for network %s", contractName, networkID)
			}
			return address, nil
		}
	}
	
	return "", fmt.Errorf("network not found: %s", networkID)
}

// SetContractAddress sets the contract address for a specific network and contract name
func (c *BaaSConfig) SetContractAddress(networkID, contractName, address string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, network := range c.Networks {
		if network.NetworkID == networkID {
			if c.Networks[i].ContractAddresses == nil {
				c.Networks[i].ContractAddresses = make(map[string]string)
			}
			c.Networks[i].ContractAddresses[contractName] = address
			return nil
		}
	}
	
	return fmt.Errorf("network not found: %s", networkID)
}

// GetAPIKey returns the API key for a specific network and service
func (c *BaaSConfig) GetAPIKey(networkID, service string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	for _, network := range c.Networks {
		if network.NetworkID == networkID {
			apiKey, ok := network.ApiKeys[service]
			if !ok {
				return "", fmt.Errorf("API key for service %s not found for network %s", service, networkID)
			}
			return apiKey, nil
		}
	}
	
	return "", fmt.Errorf("network not found: %s", networkID)
}

// SetAPIKey sets the API key for a specific network and service
func (c *BaaSConfig) SetAPIKey(networkID, service, apiKey string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for i, network := range c.Networks {
		if network.NetworkID == networkID {
			if c.Networks[i].ApiKeys == nil {
				c.Networks[i].ApiKeys = make(map[string]string)
			}
			c.Networks[i].ApiKeys[service] = apiKey
			return nil
		}
	}
	
	return fmt.Errorf("network not found: %s", networkID)
}

// GetBridgesByNetworkPair returns all bridges between two networks
func (c *BaaSConfig) GetBridgesByNetworkPair(sourceNetworkID, destinationNetworkID string) []BridgeConfiguration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	bridges := []BridgeConfiguration{}
	for _, bridge := range c.CrossChainConfig.BridgeConfigurations {
		if bridge.SourceNetworkID == sourceNetworkID && bridge.DestinationNetworkID == destinationNetworkID && bridge.Enabled {
			bridges = append(bridges, bridge)
		}
	}
	
	return bridges
}

// GetAssetMappingsByChain returns all asset mappings for a specific chain
func (c *BaaSConfig) GetAssetMappingsByChain(chainID string) []AssetMapping {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	mappings := []AssetMapping{}
	for _, mapping := range c.CrossChainConfig.AssetMappings {
		for _, chainMapping := range mapping.ChainMappings {
			if chainMapping.ChainID == chainID {
				mappings = append(mappings, mapping)
				break
			}
		}
	}
	
	return mappings
}

// CreateDefaultConfig creates a default BaaS configuration
func CreateDefaultConfig() *BaaSConfig {
	return &BaaSConfig{
		ServiceID:      "baas-default",
		ServiceName:    "BaaS Service",
		ServiceVersion: "1.0.0",
		LogLevel:       "info",
		APIKey:         "",
		APIEndpoint:    "http://blockchain-mock:8545",
		Networks: []NetworkConfig{
			{
				NetworkID:          "ethereum-mainnet",
				NetworkName:        "Ethereum Mainnet",
				NetworkType:        "ethereum",
				NetworkVersion:     "1.0.0",
				Endpoints:          []string{"https://mainnet.infura.io/v3/YOUR_API_KEY"},
				ChainID:            "1",
				ConsensusAlgorithm: "PoW",
				NetworkParams:      map[string]interface{}{},
				ContractAddresses:  map[string]string{},
				ApiKeys:            map[string]string{},
				GasPrice:           "auto",
				GasLimit:           3000000,
				Enabled:            true,
			},
			{
				NetworkID:          "cosmos-hub",
				NetworkName:        "Cosmos Hub",
				NetworkType:        "cosmos",
				NetworkVersion:     "1.0.0",
				Endpoints:          []string{"https://rpc.cosmos.network"},
				ChainID:            "cosmoshub-4",
				ConsensusAlgorithm: "Tendermint",
				NetworkParams:      map[string]interface{}{},
				ContractAddresses:  map[string]string{},
				ApiKeys:            map[string]string{},
				Enabled:            true,
			},
			{
				NetworkID:          "polkadot-mainnet",
				NetworkName:        "Polkadot",
				NetworkType:        "polkadot",
				NetworkVersion:     "1.0.0",
				Endpoints:          []string{"wss://rpc.polkadot.io"},
				ChainID:            "polkadot",
				ConsensusAlgorithm: "GRANDPA",
				NetworkParams:      map[string]interface{}{},
				ContractAddresses:  map[string]string{},
				ApiKeys:            map[string]string{},
				Enabled:            true,
			},
		},
		CrossChainConfig: CrossChainConfig{
			Enabled: true,
			BridgeConfigurations: []BridgeConfiguration{
				{
					BridgeID:            "cosmos-ibc-1",
					BridgeName:          "Cosmos IBC Bridge",
					BridgeType:          "ibc",
					SourceNetworkID:     "cosmos-hub",
					DestinationNetworkID: "osmosis-1",
					ContractAddresses:   map[string]string{},
					EndpointURLs:        []string{"https://rpc.cosmos.network", "https://rpc.osmosis.zone"},
					AdditionalParams:    map[string]interface{}{
						"channelId": "channel-0",
						"portId":    "transfer",
					},
					Enabled:          true,
					VerificationMode: "optimistic",
				},
				{
					BridgeID:            "polkadot-xcm-1",
					BridgeName:          "Polkadot XCM Bridge",
					BridgeType:          "xcm",
					SourceNetworkID:     "polkadot-mainnet",
					DestinationNetworkID: "kusama-mainnet",
					ContractAddresses:   map[string]string{},
					EndpointURLs:        []string{"wss://rpc.polkadot.io", "wss://kusama-rpc.polkadot.io"},
					AdditionalParams:    map[string]interface{}{
						"relayChain": "polkadot",
						"paraId":     "1000",
					},
					Enabled:          true,
					VerificationMode: "optimistic",
				},
			},
			InteroperabilityConfig: map[string]interface{}{},
			SupportedProtocols:     []string{"IBC", "XCM", "Hash Time Lock"},
			AssetMappings:          []AssetMapping{},
			TrustedRelayers:        []string{},
			OracleURLs:             []string{},
			ValidationThreshold:    2,
			MaxPacketSize:          1048576,
			MaxTimeoutHeight:       10000,
			RouteCache:             map[string]interface{}{},
		},
		DeploymentConfig: DeploymentConfig{
			Environment:      "development",
			MinimumNodes:     1,
			RecommendedNodes: 3,
			HighAvailability: false,
			ResourceLimits:   map[string]string{
				"cpu":    "2",
				"memory": "4Gi",
				"disk":   "100Gi",
			},
			AutoScaling:       false,
			DeploymentRegions: []string{"us-west"},
		},
		SecurityConfig: SecurityConfig{
			RateLimiting:         true,
			MaxRequestsPerMinute: 60,
			IPWhitelist:          []string{},
			IPBlacklist:          []string{},
			AuthMechanisms:       []string{"jwt", "apikey"},
			CORSOrigins:          []string{"*"},
			AllowedActions:       map[string]string{},
			AuditLogging:         true,
			TLSEnabled:           true,
		},
		MonitoringConfig: MonitoringConfig{
			Enabled:             true,
			LoggingLevel:        "info",
			MetricsEnabled:      true,
			TracingEnabled:      false,
			HealthCheckInterval: 60,
			PerformanceThresholds: map[string]interface{}{
				"latency":    1000,
				"throughput": 100,
			},
		},
		IPFSConfig: IPFSConfig{
			Enabled:           true,
			APIEndpoint:       "http://localhost:5001",
			GatewayEndpoint:   "http://localhost:8080",
			ReplicationFactor: 3,
			MaxFileSize:       104857600, // 100MB
			AllowedMimeTypes:  []string{"application/json", "text/plain", "image/png", "image/jpeg", "application/pdf"},
			EncryptionEnabled: false,
			AuthType:          "none",
		},
		InfrastructureConfig: InfrastructureConfig{
			ResourceRequirements: map[string]string{
				"cpu":    "2",
				"memory": "4Gi",
				"disk":   "100Gi",
			},
			StorageSize:            "100Gi",
			BackupEnabled:          true,
			BackupRetentionDays:    7,
			DisasterRecoveryEnabled: false,
			HAEnabled:              false,
		},
		APIConfig: APIConfig{
			RestEnabled:        true,
			RestPort:           8000,
			GraphQLEnabled:     false,
			GraphQLPort:        8001,
			WebsocketEnabled:   true,
			WebsocketPort:      8002,
			RateLimit:          60,
			DocEnabled:         true,
			DocPath:            "/docs",
			HealthCheckPath:    "/health",
			RequestTimeout:     30,
			MaxRequestBodySize: 10485760, // 10MB
			EnableCompression:  true,
			AllowedOrigins:     []string{"*"},
			AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:     []string{"Content-Type", "Authorization", "X-API-Key"},
			ExposedHeaders:     []string{},
			AllowCredentials:   false,
		},
		FallbackServices: []FallbackService{},
		GovernanceConfig: GovernanceConfig{
			VotingEnabled:        false,
			ProposalThreshold:    10,
			VotingPeriod:         7200,
			VotingQuorum:         51,
			EnactmentDelay:       1440,
			ProposalTypes:        []string{"parameter", "upgrade", "asset"},
			VetoThreshold:        33,
			EmergencyVetoEnabled: true,
		},
	}
}