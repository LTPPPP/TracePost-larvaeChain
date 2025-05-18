package analytics

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// SystemMetrics represents overall system performance metrics
type SystemMetrics struct {
	ActiveUsers         int       `json:"active_users"`
	TotalBatches        int       `json:"total_batches"`
	BlockchainTxCount   int       `json:"blockchain_tx_count"`
	APIRequestsPerHour  int       `json:"api_requests_per_hour"`
	AvgResponseTime     float64   `json:"avg_response_time_ms"`
	SystemHealth        string    `json:"system_health"` // "healthy", "degraded", "issues"
	ServerCPUUsage      float64   `json:"server_cpu_usage"`
	ServerMemoryUsage   float64   `json:"server_memory_usage"`
	DbConnections       int       `json:"db_connections"`
	LastUpdated         time.Time `json:"last_updated"`
}

// ComplianceMetrics represents compliance-related metrics
type ComplianceMetrics struct {
	TotalCertificates     int                   `json:"total_certificates"`
	ValidCertificates     int                   `json:"valid_certificates"`
	ExpiredCertificates   int                   `json:"expired_certificates"`
	RevokedCertificates   int                   `json:"revoked_certificates"`
	CompanyCompliance     map[string]float64    `json:"company_compliance"` // company name -> compliance percentage
	StandardsCompliance   map[string]float64    `json:"standards_compliance"` // standard name -> compliance percentage
	RegionalCompliance    map[string]float64    `json:"regional_compliance"` // region -> compliance percentage
	ComplianceTrends      map[string][]float64  `json:"compliance_trends"` // period -> compliance values
	LastUpdated           time.Time             `json:"last_updated"`
}

// BlockchainMetrics represents blockchain-related metrics
type BlockchainMetrics struct {
	TotalNodes              int                       `json:"total_nodes"`
	ActiveNodes             int                       `json:"active_nodes"`
	NetworkHealth           string                    `json:"network_health"` // "healthy", "degraded", "issues"
	ConsensusStatus         string                    `json:"consensus_status"` // "running", "syncing", "stalled"
	AverageBlockTime        float64                   `json:"average_block_time_ms"`
	TransactionsPerSecond   float64                   `json:"transactions_per_second"`
	PendingTransactions     int                       `json:"pending_transactions"`
	NodeLatencies           map[string]int            `json:"node_latencies"` // node id -> latency in ms
	ChainHealth             map[string]string         `json:"chain_health"` // chain id -> health status
	CrossChainTransactions  map[string]int            `json:"cross_chain_transactions"` // source-target chain -> count
	LastUpdated             time.Time                 `json:"last_updated"`
}

// UserActivityMetrics represents user activity metrics
type UserActivityMetrics struct {
	ActiveUsersByRole      map[string]int         `json:"active_users_by_role"` // role -> count
	LoginFrequency         map[string]int         `json:"login_frequency"` // period -> count
	APIEndpointUsage       map[string]int         `json:"api_endpoint_usage"` // endpoint -> count
	MostActiveUsers        []models.UserActivity  `json:"most_active_users"`
	UserGrowth             map[string]int         `json:"user_growth"` // period -> new users
	LastUpdated            time.Time              `json:"last_updated"`
}

// BatchMetrics represents metrics related to batch production and tracking
type BatchMetrics struct {
	TotalBatchesProduced   int                       `json:"total_batches_produced"`
	ActiveBatches          int                       `json:"active_batches"`
	BatchesByStatus        map[string]int            `json:"batches_by_status"` // status -> count
	BatchesByRegion        map[string]int            `json:"batches_by_region"` // region -> count
	BatchesBySpecies       map[string]int            `json:"batches_by_species"` // species -> count
	BatchesByHatchery      map[string]int            `json:"batches_by_hatchery"` // hatchery -> count
	ProductionTrend        map[string][]int          `json:"production_trend"` // period -> production values
	AverageShipmentTime    map[string]float64        `json:"average_shipment_time"` // route -> avg time in hours
	LastUpdated            time.Time                 `json:"last_updated"`
}

// AnalyticsService provides analytics data collection and aggregation
type AnalyticsService struct {
	mutex             sync.RWMutex
	systemMetrics     SystemMetrics
	complianceMetrics ComplianceMetrics
	blockchainMetrics BlockchainMetrics
	userActivityMetrics UserActivityMetrics
	batchMetrics      BatchMetrics
	updateInterval    time.Duration
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService() *AnalyticsService {
	service := &AnalyticsService{
		updateInterval: 5 * time.Minute,
	}
	
	// Initialize metrics with empty maps to avoid nil map errors
	service.complianceMetrics.CompanyCompliance = make(map[string]float64)
	service.complianceMetrics.StandardsCompliance = make(map[string]float64)
	service.complianceMetrics.RegionalCompliance = make(map[string]float64)
	service.complianceMetrics.ComplianceTrends = make(map[string][]float64)
	
	service.blockchainMetrics.NodeLatencies = make(map[string]int)
	service.blockchainMetrics.ChainHealth = make(map[string]string)
	service.blockchainMetrics.CrossChainTransactions = make(map[string]int)
	
	service.userActivityMetrics.ActiveUsersByRole = make(map[string]int)
	service.userActivityMetrics.LoginFrequency = make(map[string]int)
	service.userActivityMetrics.APIEndpointUsage = make(map[string]int)
	service.userActivityMetrics.MostActiveUsers = make([]models.UserActivity, 0)
	service.userActivityMetrics.UserGrowth = make(map[string]int)
	
	service.batchMetrics.BatchesByStatus = make(map[string]int)
	service.batchMetrics.BatchesByRegion = make(map[string]int)
	service.batchMetrics.BatchesBySpecies = make(map[string]int)
	service.batchMetrics.BatchesByHatchery = make(map[string]int)
	service.batchMetrics.ProductionTrend = make(map[string][]int)
	service.batchMetrics.AverageShipmentTime = make(map[string]float64)
	
	return service
}

// StartCollector starts the analytics data collection process
func (as *AnalyticsService) StartCollector() {
	go func() {
		// Initial collection
		as.CollectAllMetrics()
		
		// Schedule regular collection
		ticker := time.NewTicker(as.updateInterval)
		for {
			select {
			case <-ticker.C:
				as.CollectAllMetrics()
			}
		}
	}()
}

// CollectAllMetrics collects all metrics from various system components
func (as *AnalyticsService) CollectAllMetrics() {
	as.CollectSystemMetrics()
	as.CollectComplianceMetrics()
	as.CollectBlockchainMetrics()
	as.CollectUserActivityMetrics()
	as.CollectBatchMetrics()
}

// CollectSystemMetrics collects system performance metrics
func (as *AnalyticsService) CollectSystemMetrics() {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	
	// Query active users
	var activeUsers int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE is_active = true`).Scan(&activeUsers)
	if err != nil {
		fmt.Println("Error querying active users:", err)
	}
	
	// Query total batches
	var totalBatches int
	err = db.DB.QueryRow(`SELECT COUNT(*) FROM batches`).Scan(&totalBatches)
	if err != nil {
		fmt.Println("Error querying total batches:", err)
	}
	
	// Query blockchain transactions
	var txCount int
	err = db.DB.QueryRow(`SELECT COUNT(*) FROM blockchain_records`).Scan(&txCount)
	if err != nil {
		fmt.Println("Error querying blockchain transactions:", err)
	}
	
	// Query API requests in the last hour
	var requestsPerHour int
	err = db.DB.QueryRow(`SELECT COUNT(*) FROM api_logs WHERE created_at > NOW() - INTERVAL '1 hour'`).Scan(&requestsPerHour)
	if err != nil {
		// If table doesn't exist or other issue, we'll just use a default value
		requestsPerHour = 0
		fmt.Println("Error querying API requests:", err)
	}
	
	// Query average response time
	var avgResponseTime float64
	err = db.DB.QueryRow(`SELECT AVG(response_time) FROM api_logs WHERE created_at > NOW() - INTERVAL '1 hour'`).Scan(&avgResponseTime)
	if err != nil {
		// If table doesn't exist or other issue, we'll just use a default value
		avgResponseTime = 0
		fmt.Println("Error querying average response time:", err)
	}
	
	// Update system metrics
	as.systemMetrics = SystemMetrics{
		ActiveUsers:         activeUsers,
		TotalBatches:        totalBatches,
		BlockchainTxCount:   txCount,
		APIRequestsPerHour:  requestsPerHour,
		AvgResponseTime:     avgResponseTime,
		SystemHealth:        "healthy", // This should be determined by thresholds
		ServerCPUUsage:      35.5,      // In a real system, this would be collected from the host
		ServerMemoryUsage:   45.2,      // In a real system, this would be collected from the host
		DbConnections:       8,         // In a real system, this would be collected from the DB
		LastUpdated:         time.Now(),
	}
}

// CollectComplianceMetrics collects compliance-related metrics
func (as *AnalyticsService) CollectComplianceMetrics() {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	
	// Initialize metrics
	metrics := ComplianceMetrics{
		CompanyCompliance:     make(map[string]float64),
		StandardsCompliance:   make(map[string]float64),
		RegionalCompliance:    make(map[string]float64),
		ComplianceTrends:      make(map[string][]float64),
		LastUpdated:           time.Now(),
	}
	
	// Query certificate counts
	err := db.DB.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'valid' AND expiry_date > NOW() THEN 1 ELSE 0 END) as valid,
			SUM(CASE WHEN status = 'valid' AND expiry_date <= NOW() THEN 1 ELSE 0 END) as expired,
			SUM(CASE WHEN status = 'revoked' THEN 1 ELSE 0 END) as revoked
		FROM documents
		WHERE document_type = 'certificate'
	`).Scan(&metrics.TotalCertificates, &metrics.ValidCertificates, &metrics.ExpiredCertificates, &metrics.RevokedCertificates)
	
	if err != nil {
		fmt.Println("Error querying certificates:", err)
	}
	
	// Query company compliance
	rows, err := db.DB.Query(`
		SELECT 
			c.name,
			COUNT(CASE WHEN d.status = 'valid' AND d.expiry_date > NOW() THEN 1 ELSE NULL END) * 100.0 / COUNT(*) as compliance_percentage
		FROM 
			companies c
		JOIN 
			hatcheries h ON c.id = h.company_id
		JOIN 
			batches b ON h.id = b.hatchery_id
		LEFT JOIN 
			documents d ON b.id = d.batch_id AND d.document_type = 'certificate'
		GROUP BY 
			c.name
	`)
	
	if err != nil {
		fmt.Println("Error querying company compliance:", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var companyName string
			var compliancePercentage float64
			if err := rows.Scan(&companyName, &compliancePercentage); err != nil {
				fmt.Println("Error scanning company compliance row:", err)
				continue
			}
			metrics.CompanyCompliance[companyName] = compliancePercentage
		}
	}
	
	// Set some sample compliance data for standards
	metrics.StandardsCompliance["ASC"] = 92.3
	metrics.StandardsCompliance["ISO9001"] = 88.7
	metrics.StandardsCompliance["GlobalG.A.P"] = 85.1
	metrics.StandardsCompliance["BAP"] = 90.5
	
	// Set some sample regional compliance data
	metrics.RegionalCompliance["North Vietnam"] = 88.2
	metrics.RegionalCompliance["Central Vietnam"] = 92.7
	metrics.RegionalCompliance["South Vietnam"] = 85.3
	
	// Set some sample compliance trends
	metrics.ComplianceTrends["last_6_months"] = []float64{83.2, 84.5, 86.1, 87.2, 88.5, 89.3}
	
	as.complianceMetrics = metrics
}

// CollectBlockchainMetrics collects blockchain network metrics
func (as *AnalyticsService) CollectBlockchainMetrics() {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	
	// Query blockchain nodes
	var totalNodes, activeNodes int
	err := db.DB.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END) as active
		FROM blockchain_nodes
	`).Scan(&totalNodes, &activeNodes)
	
	if err != nil {
		// If table doesn't exist, use default values
		totalNodes = 5
		activeNodes = 5
		fmt.Println("Error querying blockchain nodes:", err)
	}
	
	// Initialize metrics with sample data (in a real system, these would come from blockchain APIs)
	metrics := BlockchainMetrics{
		TotalNodes:            totalNodes,
		ActiveNodes:           activeNodes,
		NetworkHealth:         "healthy",
		ConsensusStatus:       "running",
		AverageBlockTime:      2500.0, // ms
		TransactionsPerSecond: 15.7,
		PendingTransactions:   23,
		NodeLatencies:         make(map[string]int),
		ChainHealth:           make(map[string]string),
		CrossChainTransactions: make(map[string]int),
		LastUpdated:           time.Now(),
	}
	
	// Sample node latencies
	metrics.NodeLatencies = map[string]int{
		"node-1": 45,  // ms
		"node-2": 62,
		"node-3": 38,
		"node-4": 72,
		"node-5": 55,
	}
	
	// Sample chain health
	metrics.ChainHealth = map[string]string{
		"tracepost-main": "healthy",
		"cosmos-ibc":     "healthy",
		"polkadot":       "syncing",
	}
	
	// Sample cross-chain transactions
	metrics.CrossChainTransactions = map[string]int{
		"tracepost-main->cosmos-ibc": 137,
		"cosmos-ibc->tracepost-main": 92,
		"tracepost-main->polkadot":   43,
		"polkadot->tracepost-main":   38,
	}
	
	as.blockchainMetrics = metrics
}

// CollectUserActivityMetrics collects user activity metrics
func (as *AnalyticsService) CollectUserActivityMetrics() {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	
	// Initialize metrics
	metrics := UserActivityMetrics{
		ActiveUsersByRole:   make(map[string]int),
		LoginFrequency:      make(map[string]int),
		APIEndpointUsage:    make(map[string]int),
		MostActiveUsers:     make([]models.UserActivity, 0),
		UserGrowth:          make(map[string]int),
		LastUpdated:         time.Now(),
	}
	
	// Query active users by role
	rows, err := db.DB.Query(`
		SELECT
			role,
			COUNT(*)
		FROM users
		WHERE is_active = true
		GROUP BY role
	`)
	
	if err != nil {
		fmt.Println("Error querying active users by role:", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var role string
			var count int
			if err := rows.Scan(&role, &count); err != nil {
				fmt.Println("Error scanning active users row:", err)
				continue
			}
			metrics.ActiveUsersByRole[role] = count
		}
	}
	
	// Set some sample login frequency data
	metrics.LoginFrequency = map[string]int{
		"today":           128,
		"yesterday":       115,
		"last_7_days":     742,
		"last_30_days":    2857,
	}
	
	// Set some sample API endpoint usage data
	metrics.APIEndpointUsage = map[string]int{
		"/api/v1/batches":             1257,
		"/api/v1/hatcheries":          892,
		"/api/v1/documents":           745,
		"/api/v1/auth/login":          684,
		"/api/v1/blockchain/batch":    523,
	}
	
	// Set some sample most active users
	metrics.MostActiveUsers = []models.UserActivity{
		{UserID: 1, Username: "admin", RequestCount: 248, LastActive: time.Now().Add(-time.Hour)},
		{UserID: 5, Username: "farm_manager1", RequestCount: 187, LastActive: time.Now().Add(-30 * time.Minute)},
		{UserID: 12, Username: "inspector_nguyen", RequestCount: 156, LastActive: time.Now().Add(-2 * time.Hour)},
		{UserID: 8, Username: "logistics_tran", RequestCount: 134, LastActive: time.Now().Add(-15 * time.Minute)},
		{UserID: 23, Username: "hatchery_director", RequestCount: 112, LastActive: time.Now().Add(-5 * time.Hour)},
	}
	
	// Set some sample user growth data
	metrics.UserGrowth = map[string]int{
		"today":           3,
		"yesterday":       2,
		"last_7_days":     15,
		"last_30_days":    42,
		"last_90_days":    98,
	}
	
	as.userActivityMetrics = metrics
}

// CollectBatchMetrics collects batch-related metrics
func (as *AnalyticsService) CollectBatchMetrics() {
	as.mutex.Lock()
	defer as.mutex.Unlock()
	
	// Query total batches and active batches
	var totalBatches, activeBatches int
	err := db.DB.QueryRow(`
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END) as active
		FROM batches
	`).Scan(&totalBatches, &activeBatches)
	
	if err != nil {
		fmt.Println("Error querying batches:", err)
	}
	
	// Initialize metrics
	metrics := BatchMetrics{
		TotalBatchesProduced: totalBatches,
		ActiveBatches:        activeBatches,
		BatchesByStatus:      make(map[string]int),
		BatchesByRegion:      make(map[string]int),
		BatchesBySpecies:     make(map[string]int),
		BatchesByHatchery:    make(map[string]int),
		ProductionTrend:      make(map[string][]int),
		AverageShipmentTime:  make(map[string]float64),
		LastUpdated:          time.Now(),
	}
	
	// Query batches by status
	rows, err := db.DB.Query(`
		SELECT
			status,
			COUNT(*)
		FROM batches
		GROUP BY status
	`)
	
	if err != nil {
		fmt.Println("Error querying batches by status:", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var status string
			var count int
			if err := rows.Scan(&status, &count); err != nil {
				fmt.Println("Error scanning batches by status row:", err)
				continue
			}
			metrics.BatchesByStatus[status] = count
		}
	}
	
	// Query batches by hatchery
	rows, err = db.DB.Query(`
		SELECT
			h.name,
			COUNT(b.id)
		FROM batches b
		JOIN hatcheries h ON b.hatchery_id = h.id
		GROUP BY h.name
	`)
	
	if err != nil {
		fmt.Println("Error querying batches by hatchery:", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var hatcheryName string
			var count int
			if err := rows.Scan(&hatcheryName, &count); err != nil {
				fmt.Println("Error scanning batches by hatchery row:", err)
				continue
			}
			metrics.BatchesByHatchery[hatcheryName] = count
		}
	}
	
	// Query batches by species
	rows, err = db.DB.Query(`
		SELECT
			species,
			COUNT(id)
		FROM batches
		GROUP BY species
	`)
	
	if err != nil {
		fmt.Println("Error querying batches by species:", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var species string
			var count int
			if err := rows.Scan(&species, &count); err != nil {
				fmt.Println("Error scanning batches by species row:", err)
				continue
			}
			metrics.BatchesBySpecies[species] = count
		}
	}
	
	// Set some sample batches by region data
	metrics.BatchesByRegion = map[string]int{
		"North Vietnam":   287,
		"Central Vietnam": 452,
		"South Vietnam":   339,
	}
	
	// Set some sample production trend data
	metrics.ProductionTrend = map[string][]int{
		"last_6_months": {125, 142, 160, 155, 172, 183},
		"last_12_months": {98, 105, 112, 125, 132, 140, 142, 160, 155, 172, 183, 192},
	}
	
	// Set some sample average shipment time data
	metrics.AverageShipmentTime = map[string]float64{
		"North to Central": 5.2,  // hours
		"Central to South": 4.8,
		"North to South":   10.5,
		"Farm to Processing": 3.2,
		"Processing to Distribution": 6.7,
	}
	
	as.batchMetrics = metrics
}

// GetSystemMetrics returns the current system metrics
func (as *AnalyticsService) GetSystemMetrics() SystemMetrics {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.systemMetrics
}

// GetComplianceMetrics returns the current compliance metrics
func (as *AnalyticsService) GetComplianceMetrics() ComplianceMetrics {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.complianceMetrics
}

// GetBlockchainMetrics returns the current blockchain metrics
func (as *AnalyticsService) GetBlockchainMetrics() BlockchainMetrics {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.blockchainMetrics
}

// GetUserActivityMetrics returns the current user activity metrics
func (as *AnalyticsService) GetUserActivityMetrics() UserActivityMetrics {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.userActivityMetrics
}

// GetBatchMetrics returns the current batch metrics
func (as *AnalyticsService) GetBatchMetrics() BatchMetrics {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return as.batchMetrics
}

// GetAllMetrics returns all analytics metrics
func (as *AnalyticsService) GetAllMetrics() map[string]interface{} {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	
	return map[string]interface{}{
		"system":        as.systemMetrics,
		"compliance":    as.complianceMetrics,
		"blockchain":    as.blockchainMetrics,
		"user_activity": as.userActivityMetrics,
		"batch":         as.batchMetrics,
		"timestamp":     time.Now(),
	}
}

// LogAPIRequest logs API request for analytics
func (as *AnalyticsService) LogAPIRequest(endpoint string, userID int, responseTime float64) {
	// In a real implementation, this would insert into a database table
	// For now, we'll just print it
	fmt.Printf("API Request: %s, User: %d, Response Time: %.2f ms\n", endpoint, userID, responseTime)
}

// GetMetricsJSON returns all metrics as a JSON string
func (as *AnalyticsService) GetMetricsJSON() (string, error) {
	metrics := as.GetAllMetrics()
	
	jsonBytes, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return "", err
	}
	
	return string(jsonBytes), nil
}

// Global instance for easy access across packages
var AnalyticsInstance *AnalyticsService
var once sync.Once

// InitAnalytics initializes the analytics service singleton
func InitAnalytics() {
	once.Do(func() {
		AnalyticsInstance = NewAnalyticsService()
		AnalyticsInstance.StartCollector()
	})
}

// GetAnalytics returns the analytics service instance
func GetAnalytics() *AnalyticsService {
	if AnalyticsInstance == nil {
		InitAnalytics()
	}
	return AnalyticsInstance
}
