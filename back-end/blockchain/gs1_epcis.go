// gs1_epcis.go
package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// EPCISClient provides integration with GS1 EPCIS standard
type EPCISClient struct {
	Config EPCISConfig
	HTTPClient *http.Client
}

// EPCISConfig contains configuration for connecting to EPCIS systems
type EPCISConfig struct {
	RESTEndpoint string
	SOAPEndpoint string
	Username     string
	Password     string
	APIKey       string
	DefaultGLN   string // Global Location Number
	CompanyPrefix string
	VersionInfo  string
	Namespaces   map[string]string
}

// EPCISEvent represents an EPCIS event
type EPCISEvent struct {
	EventType     string    // ObjectEvent, AggregationEvent, TransactionEvent, TransformationEvent
	EventTime     time.Time
	RecordTime    time.Time
	EventTimeZone string
	EPCs          []string  // Electronic Product Codes
	Action        string    // ADD, OBSERVE, DELETE
	BizStep       string    // EPCIS business step (e.g., shipping, receiving)
	Disposition   string    // EPCIS disposition (e.g., in_transit, in_progress)
	ReadPoint     string    // GLN of the location where the event occurred
	BizLocation   string    // GLN of the business location
	BizTransactions []EPCISBizTransaction
	SourceList    []EPCISSource
	DestinationList []EPCISDestination
	ILMDs         map[string]interface{} // Instance/Lot Master Data
	Extensions    map[string]interface{}
}

// EPCISBizTransaction represents a business transaction in EPCIS
type EPCISBizTransaction struct {
	Type string // Purchase Order, Despatch Advice, etc.
	ID   string
}

// EPCISSource represents a source in EPCIS
type EPCISSource struct {
	Type string // Owning Party, Possessing Party, etc.
	ID   string
}

// EPCISDestination represents a destination in EPCIS
type EPCISDestination struct {
	Type string // Owning Party, Possessing Party, etc.
	ID   string
}

// EPCISDocument represents an EPCIS document
type EPCISDocument struct {
	SchemaVersion      string
	CreationDate       time.Time
	EPCISBody          EPCISBody
}

// EPCISBody represents the body of an EPCIS document
type EPCISBody struct {
	EventList          []EPCISEvent
}

// EPCISQuery represents a query for EPCIS events
type EPCISQuery struct {
	QueryName    string
	Parameters   map[string]interface{}
}

// EPCISQueryResult represents the result of an EPCIS query
type EPCISQueryResult struct {
	QueryName    string
	ResultCount  int
	Events       []EPCISEvent
}

// NewEPCISClient creates a new EPCIS client
func NewEPCISClient(config EPCISConfig) *EPCISClient {
	// Set default namespaces if not provided
	if config.Namespaces == nil {
		config.Namespaces = map[string]string{
			"epcis": "urn:epcglobal:epcis:xsd:1",
			"sbdh":  "http://www.unece.org/cefact/namespaces/StandardBusinessDocumentHeader",
		}
	}
	
	return &EPCISClient{
		Config: config,
		HTTPClient: &http.Client{
			Timeout: time.Duration(30) * time.Second,
		},
	}
}

// ConvertBatchToEPCISEvent converts a batch to an EPCIS event
func (ec *EPCISClient) ConvertBatchToEPCISEvent(batch map[string]interface{}) (*EPCISEvent, error) {
	// Extract batch information
	batchID, ok := batch["batch_id"].(string)
	if !ok {
		return nil, errors.New("batch_id not found or not a string")
	}
	
	// Create EPC based on batch ID
	epc := fmt.Sprintf("urn:epc:id:sgtin:%s.%s", ec.Config.CompanyPrefix, batchID)
	
	// Extract other batch information
	status, _ := batch["status"].(string)
	createdAt, _ := batch["created_at"].(time.Time)
	hatcheryID, _ := batch["hatchery_id"].(string)
	
	// Determine EPCIS business step based on status
	bizStep := "urn:epcglobal:cbv:bizstep:commissioning" // Default
	if status == "shipped" {
		bizStep = "urn:epcglobal:cbv:bizstep:shipping"
	} else if status == "received" {
		bizStep = "urn:epcglobal:cbv:bizstep:receiving"
	}
	
	// Determine EPCIS disposition based on status
	disposition := "urn:epcglobal:cbv:disp:active" // Default
	if status == "shipped" {
		disposition = "urn:epcglobal:cbv:disp:in_transit"
	} else if status == "completed" {
		disposition = "urn:epcglobal:cbv:disp:sold"
	}
	
	// Create EPCIS event
	event := &EPCISEvent{
		EventType:     "ObjectEvent",
		EventTime:     createdAt,
		RecordTime:    time.Now(),
		EventTimeZone: "+07:00", // Vietnam time zone
		EPCs:          []string{epc},
		Action:        "ADD",
		BizStep:       bizStep,
		Disposition:   disposition,
		ReadPoint:     fmt.Sprintf("urn:epc:id:sgln:%s.%s", ec.Config.CompanyPrefix, hatcheryID),
		BizLocation:   fmt.Sprintf("urn:epc:id:sgln:%s.%s", ec.Config.CompanyPrefix, hatcheryID),
		BizTransactions: []EPCISBizTransaction{
			{
				Type: "urn:epcglobal:cbv:btt:po",
				ID:   "urn:epcglobal:cbv:bt:" + batchID,
			},
		},
		ILMDs: map[string]interface{}{
			"aquaculture:species": batch["species"],
			"aquaculture:quantity": batch["quantity"],
		},
		Extensions: map[string]interface{}{
			"tracepost:version": "1.0",
		},
	}
	
	return event, nil
}

// CreateEPCISDocument creates an EPCIS document from a batch
func (ec *EPCISClient) CreateEPCISDocument(batch map[string]interface{}) (*EPCISDocument, error) {
	// Convert batch to EPCIS event
	event, err := ec.ConvertBatchToEPCISEvent(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to convert batch to EPCIS event: %w", err)
	}
	
	// Create EPCIS document
	document := &EPCISDocument{
		SchemaVersion: "1.2",
		CreationDate:  time.Now(),
		EPCISBody: EPCISBody{
			EventList: []EPCISEvent{*event},
		},
	}
	
	return document, nil
}

// PublishEPCISDocument publishes an EPCIS document to an EPCIS repository
func (ec *EPCISClient) PublishEPCISDocument(document *EPCISDocument) error {
	// Convert document to XML
	documentXML, err := ec.ConvertDocumentToXML(document)
	if err != nil {
		return fmt.Errorf("failed to convert document to XML: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", ec.Config.RESTEndpoint, bytes.NewBuffer([]byte(documentXML)))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/xml")
	if ec.Config.APIKey != "" {
		req.Header.Set("Authorization", "ApiKey "+ec.Config.APIKey)
	}
	
	// Send request
	resp, err := ec.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to publish EPCIS document, status code: %d", resp.StatusCode)
	}
	
	return nil
}

// ConvertDocumentToXML converts an EPCIS document to XML
func (ec *EPCISClient) ConvertDocumentToXML(document *EPCISDocument) (string, error) {
	// This is a simplified implementation, in a real system this would use proper XML marshalling
	// with all the required EPCIS namespaces and XML structure
	
	// Build XML header
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	xmlBuilder.WriteString(`<epcis:EPCISDocument xmlns:epcis="urn:epcglobal:epcis:xsd:1" schemaVersion="1.2" creationDate="`)
	xmlBuilder.WriteString(document.CreationDate.Format(time.RFC3339))
	xmlBuilder.WriteString(`">`)
	
	// EPCISBody
	xmlBuilder.WriteString(`<EPCISBody>`)
	xmlBuilder.WriteString(`<EventList>`)
	
	// Process each event
	for _, event := range document.EPCISBody.EventList {
		// ObjectEvent
		xmlBuilder.WriteString(`<ObjectEvent>`)
		
		// Event time
		xmlBuilder.WriteString(`<eventTime>`)
		xmlBuilder.WriteString(event.EventTime.Format(time.RFC3339))
		xmlBuilder.WriteString(`</eventTime>`)
		
		// Record time
		xmlBuilder.WriteString(`<recordTime>`)
		xmlBuilder.WriteString(event.RecordTime.Format(time.RFC3339))
		xmlBuilder.WriteString(`</recordTime>`)
		
		// Event time zone offset
		xmlBuilder.WriteString(`<eventTimeZoneOffset>`)
		xmlBuilder.WriteString(event.EventTimeZone)
		xmlBuilder.WriteString(`</eventTimeZoneOffset>`)
		
		// EPCs
		xmlBuilder.WriteString(`<epcList>`)
		for _, epc := range event.EPCs {
			xmlBuilder.WriteString(`<epc>`)
			xmlBuilder.WriteString(epc)
			xmlBuilder.WriteString(`</epc>`)
		}
		xmlBuilder.WriteString(`</epcList>`)
		
		// Action
		xmlBuilder.WriteString(`<action>`)
		xmlBuilder.WriteString(event.Action)
		xmlBuilder.WriteString(`</action>`)
		
		// BizStep
		xmlBuilder.WriteString(`<bizStep>`)
		xmlBuilder.WriteString(event.BizStep)
		xmlBuilder.WriteString(`</bizStep>`)
		
		// Disposition
		xmlBuilder.WriteString(`<disposition>`)
		xmlBuilder.WriteString(event.Disposition)
		xmlBuilder.WriteString(`</disposition>`)
		
		// ReadPoint
		xmlBuilder.WriteString(`<readPoint><id>`)
		xmlBuilder.WriteString(event.ReadPoint)
		xmlBuilder.WriteString(`</id></readPoint>`)
		
		// BizLocation
		xmlBuilder.WriteString(`<bizLocation><id>`)
		xmlBuilder.WriteString(event.BizLocation)
		xmlBuilder.WriteString(`</id></bizLocation>`)
		
		// BizTransactionList
		if len(event.BizTransactions) > 0 {
			xmlBuilder.WriteString(`<bizTransactionList>`)
			for _, bizTransaction := range event.BizTransactions {
				xmlBuilder.WriteString(`<bizTransaction type="`)
				xmlBuilder.WriteString(bizTransaction.Type)
				xmlBuilder.WriteString(`">`)
				xmlBuilder.WriteString(bizTransaction.ID)
				xmlBuilder.WriteString(`</bizTransaction>`)
			}
			xmlBuilder.WriteString(`</bizTransactionList>`)
		}
		
		// End ObjectEvent
		xmlBuilder.WriteString(`</ObjectEvent>`)
	}
	
	// End EventList and EPCISBody
	xmlBuilder.WriteString(`</EventList>`)
	xmlBuilder.WriteString(`</EPCISBody>`)
	
	// End EPCISDocument
	xmlBuilder.WriteString(`</epcis:EPCISDocument>`)
	
	return xmlBuilder.String(), nil
}

// QueryEPCISRepository queries an EPCIS repository for events
func (ec *EPCISClient) QueryEPCISRepository(query EPCISQuery) (*EPCISQueryResult, error) {
	// Build query URL
	url := ec.Config.RESTEndpoint + "/query"
	
	// Convert query to JSON
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if ec.Config.APIKey != "" {
		req.Header.Set("Authorization", "ApiKey "+ec.Config.APIKey)
	}
	
	// Send request
	resp, err := ec.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status code: %d", resp.StatusCode)
	}
	
	// Parse response
	var queryResult EPCISQueryResult
	err = json.NewDecoder(resp.Body).Decode(&queryResult)
	if err != nil {
		return nil, fmt.Errorf("failed to decode query result: %w", err)
	}
	
	return &queryResult, nil
}

// TrackEPC tracks an EPC through the supply chain
func (ec *EPCISClient) TrackEPC(epc string) ([]EPCISEvent, error) {
	// Create query
	query := EPCISQuery{
		QueryName: "SimpleEventQuery",
		Parameters: map[string]interface{}{
			"MATCH_epc": []string{epc},
		},
	}
	
	// Execute query
	result, err := ec.QueryEPCISRepository(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query EPCIS repository: %w", err)
	}
	
	return result.Events, nil
}

// ExportBatchToEPCIS exports a batch to an EPCIS repository
func (ec *EPCISClient) ExportBatchToEPCIS(batch map[string]interface{}) error {
	// Create EPCIS document
	document, err := ec.CreateEPCISDocument(batch)
	if err != nil {
		return fmt.Errorf("failed to create EPCIS document: %w", err)
	}
	
	// Publish document
	err = ec.PublishEPCISDocument(document)
	if err != nil {
		return fmt.Errorf("failed to publish EPCIS document: %w", err)
	}
	
	return nil
}
