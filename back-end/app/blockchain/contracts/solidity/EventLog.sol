// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title EventLog
 * @dev Stores and tracks supply chain events and documents
 */
contract EventLog {
    struct Event {
        string shipmentId;
        string eventType;
        string dataHash;
        uint256 timestamp;
        string metadata;
        address logger;
    }

    struct Document {
        uint256 timestamp;
        string metadata;
        address logger;
    }

    // Mapping from eventId to Event
    mapping(string => Event) public events;

    // Mapping from documentHash to Document
    mapping(string => Document) public documents;

    // Events
    event EventLogged(
        string indexed shipmentId,
        string indexed eventId,
        string eventType
    );

    event DocumentLogged(string indexed documentHash, string documentId);

    /**
     * @dev Contract constructor
     */
    constructor() {
        // Initialize contract
    }

    /**
     * @dev Log a supply chain event
     * @param shipmentId Shipment identifier
     * @param eventId Unique event identifier
     * @param eventType Type of event
     * @param dataHash Hash of the event data
     * @param metadata Additional metadata (JSON string)
     */
    function logEvent(
        string memory shipmentId,
        string memory eventId,
        string memory eventType,
        string memory dataHash,
        string memory metadata
    ) public {
        // Ensure event doesn't already exist
        require(events[eventId].timestamp == 0, "Event already exists");

        // Create new event
        events[eventId] = Event({
            shipmentId: shipmentId,
            eventType: eventType,
            dataHash: dataHash,
            timestamp: block.timestamp,
            metadata: metadata,
            logger: msg.sender
        });

        // Emit event
        emit EventLogged(shipmentId, eventId, eventType);
    }

    /**
     * @dev Log a document
     * @param documentId Document identifier
     * @param documentHash Hash of the document
     * @param metadata Additional metadata (JSON string)
     */
    function logDocument(
        string memory documentId,
        string memory documentHash,
        string memory metadata
    ) public {
        // Ensure document doesn't already exist
        require(
            documents[documentHash].timestamp == 0,
            "Document already exists"
        );

        // Create new document
        documents[documentHash] = Document({
            timestamp: block.timestamp,
            metadata: metadata,
            logger: msg.sender
        });

        // Emit event
        emit DocumentLogged(documentHash, documentId);
    }

    /**
     * @dev Get event details
     * @param eventId Event identifier
     * @return shipmentId The associated shipment ID
     * @return eventType The type of event
     * @return dataHash Hash of the event data
     * @return timestamp When the event was logged
     * @return metadata Additional metadata
     */
    function getEvent(
        string memory eventId
    )
        public
        view
        returns (
            string memory shipmentId,
            string memory eventType,
            string memory dataHash,
            uint256 timestamp,
            string memory metadata
        )
    {
        // Ensure event exists
        require(events[eventId].timestamp > 0, "Event does not exist");

        Event storage event_ = events[eventId];
        return (
            event_.shipmentId,
            event_.eventType,
            event_.dataHash,
            event_.timestamp,
            event_.metadata
        );
    }

    /**
     * @dev Get document details
     * @param documentHash Document hash
     * @return timestamp When the document was logged
     * @return metadata Additional metadata
     */
    function getDocument(
        string memory documentHash
    ) public view returns (uint256 timestamp, string memory metadata) {
        // Ensure document exists
        require(
            documents[documentHash].timestamp > 0,
            "Document does not exist"
        );

        Document storage document = documents[documentHash];
        return (document.timestamp, document.metadata);
    }
}
