// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/math/SafeMath.sol";

/**
 * @title LogisticsTraceability
 * @dev Smart contract for managing logistics traceability with blockchain interoperability and DID support
 * @custom:experimental This is an experimental contract for TracePost-larvaeChain
 */
contract LogisticsTraceability is AccessControl, Pausable, Initializable {
    using SafeMath for uint256;
    using ECDSA for bytes32;

    // Role definitions
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant HATCHERY_ROLE = keccak256("HATCHERY_ROLE");
    bytes32 public constant FARM_ROLE = keccak256("FARM_ROLE");
    bytes32 public constant PROCESSOR_ROLE = keccak256("PROCESSOR_ROLE");
    bytes32 public constant CERTIFIER_ROLE = keccak256("CERTIFIER_ROLE");
    bytes32 public constant RELAY_ROLE = keccak256("RELAY_ROLE"); // For cross-chain operations

    // Cross-chain connector interface for interoperability
    ICrossChainConnector public crossChainConnector;

    // DID Registry
    mapping(string => DID) public didRegistry; // did -> DID struct
    mapping(string => string) public didDocuments; // didHash -> IPFS hash of DID document
    mapping(string => mapping(address => bool)) public didControllers; // didHash -> controller -> isController

    // Batch data structures
    mapping(string => Batch) public batches; // batchId -> Batch
    mapping(string => Event[]) public batchEvents; // batchId -> Events[]
    mapping(string => Document[]) public batchDocuments; // batchId -> Documents[]
    mapping(string => EnvironmentData[]) public batchEnvironmentData; // batchId -> EnvironmentData[]

    // Cross-chain references
    mapping(string => CrossChainReference[]) public crossChainReferences; // batchId -> CrossChainReference[]

    // Claim registry
    mapping(string => Claim) public claims; // claimId -> Claim
    mapping(string => string[]) public subjectClaims; // subject DID -> claimIds[]

    // Structs
    struct DID {
        string id;
        address[] controllers;
        uint256 created;
        uint256 updated;
        bool active;
        string metadataHash; // IPFS hash for metadata
    }

    struct Claim {
        string id;
        string issuer;
        string subject;
        string claimType;
        uint256 issuedAt;
        uint256 expiresAt;
        string dataHash; // IPFS hash for claim data
        bool revoked;
    }

    struct Batch {
        string batchId;
        string hatcheryId;
        string species;
        uint256 quantity;
        string status;
        uint256 createdAt;
        string metadataHash; // IPFS hash for additional metadata
        bool active;
    }

    struct Event {
        string batchId;
        string eventType;
        string location;
        string actorId;
        uint256 timestamp;
        string metadataHash; // IPFS hash for event details
    }

    struct Document {
        string batchId;
        string documentType;
        string ipfsHash;
        string issuer;
        uint256 timestamp;
        bool verified;
    }

    struct EnvironmentData {
        string batchId;
        uint256 timestamp;
        int256 temperature; // Scaled by 100 (e.g., 25.75°C is stored as 2575)
        int256 ph; // Scaled by 100 (e.g., 7.85 is stored as 785)
        int256 salinity; // Scaled by 100 (e.g., the salinity of 35.5 ppt is stored as 3550)
        int256 dissolvedOxygen; // Scaled by 100 (e.g., 6.75 mg/L is stored as 675)
        string metadataHash; // IPFS hash for other parameters
    }

    struct CrossChainReference {
        string sourceBatchId;
        string targetChain;
        string targetTxHash;
        string targetBatchId;
        string dataStandard;
        uint256 timestamp;
        bool verified;
    }

    // Events
    event BatchCreated(
        string batchId,
        string hatcheryId,
        string species,
        uint256 quantity
    );
    event BatchStatusUpdated(string batchId, string status);
    event EventRecorded(
        string batchId,
        string eventType,
        string location,
        string actorId
    );
    event DocumentRecorded(
        string batchId,
        string documentType,
        string ipfsHash,
        string issuer
    );
    event EnvironmentDataRecorded(
        string batchId,
        int256 temperature,
        int256 ph,
        int256 salinity,
        int256 dissolvedOxygen
    );
    event CrossChainReferenceCreated(
        string sourceBatchId,
        string targetChain,
        string targetTxHash
    );

    // DID events
    event DIDCreated(string did, address controller);
    event DIDControllerAdded(string did, address controller);
    event DIDControllerRemoved(string did, address controller);
    event DIDDeactivated(string did);
    event DIDReactivated(string did);

    // Claim events
    event ClaimIssued(
        string claimId,
        string issuer,
        string subject,
        string claimType,
        uint256 expiresAt
    );
    event ClaimRevoked(string claimId, string issuer);

    /**
     * @dev Constructor initializes the contract with default admin
     */
    constructor() {
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
    }

    /**
     * @dev Initialize the contract with cross-chain connector
     * @param _crossChainConnector Address of the cross-chain connector contract
     */
    function initialize(address _crossChainConnector) public initializer {
        require(
            _crossChainConnector != address(0),
            "Invalid connector address"
        );
        crossChainConnector = ICrossChainConnector(_crossChainConnector);
    }

    /**
     * @dev Pause contract functionality
     */
    function pause() public onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /**
     * @dev Unpause contract functionality
     */
    function unpause() public onlyRole(ADMIN_ROLE) {
        _unpause();
    }

    // ========== BATCH OPERATIONS ==========

    /**
     * @dev Create a new batch
     * @param _batchId Unique identifier for the batch
     * @param _hatcheryId ID of the hatchery creating the batch
     * @param _species Species of shrimp in the batch
     * @param _quantity Quantity of shrimp larvae in the batch
     * @param _metadataHash IPFS hash to additional metadata
     */
    function createBatch(
        string memory _batchId,
        string memory _hatcheryId,
        string memory _species,
        uint256 _quantity,
        string memory _metadataHash
    ) public whenNotPaused onlyRole(HATCHERY_ROLE) {
        require(
            bytes(batches[_batchId].batchId).length == 0,
            "Batch already exists"
        );

        Batch memory newBatch = Batch({
            batchId: _batchId,
            hatcheryId: _hatcheryId,
            species: _species,
            quantity: _quantity,
            status: "created",
            createdAt: block.timestamp,
            metadataHash: _metadataHash,
            active: true
        });

        batches[_batchId] = newBatch;

        emit BatchCreated(_batchId, _hatcheryId, _species, _quantity);
    }

    /**
     * @dev Update batch status
     * @param _batchId Batch ID to update
     * @param _status New status
     */
    function updateBatchStatus(
        string memory _batchId,
        string memory _status
    ) public whenNotPaused {
        require(bytes(batches[_batchId].batchId).length > 0, "Batch not found");
        require(batches[_batchId].active, "Batch is not active");

        // Only certain roles can update to certain statuses
        if (
            keccak256(abi.encodePacked(_status)) ==
            keccak256(abi.encodePacked("processing"))
        ) {
            require(
                hasRole(PROCESSOR_ROLE, msg.sender),
                "Must have processor role"
            );
        } else if (
            keccak256(abi.encodePacked(_status)) ==
            keccak256(abi.encodePacked("certified"))
        ) {
            require(
                hasRole(CERTIFIER_ROLE, msg.sender),
                "Must have certifier role"
            );
        } else {
            require(
                hasRole(HATCHERY_ROLE, msg.sender) ||
                    hasRole(FARM_ROLE, msg.sender) ||
                    hasRole(ADMIN_ROLE, msg.sender),
                "Not authorized to update status"
            );
        }

        batches[_batchId].status = _status;

        emit BatchStatusUpdated(_batchId, _status);
    }

    /**
     * @dev Record event for a batch
     * @param _batchId Batch ID
     * @param _eventType Type of event
     * @param _location Location where event occurred
     * @param _actorId ID of the actor recording the event
     * @param _metadataHash IPFS hash for event details
     */
    function recordEvent(
        string memory _batchId,
        string memory _eventType,
        string memory _location,
        string memory _actorId,
        string memory _metadataHash
    ) public whenNotPaused {
        require(bytes(batches[_batchId].batchId).length > 0, "Batch not found");
        require(batches[_batchId].active, "Batch is not active");

        // Create event
        Event memory newEvent = Event({
            batchId: _batchId,
            eventType: _eventType,
            location: _location,
            actorId: _actorId,
            timestamp: block.timestamp,
            metadataHash: _metadataHash
        });

        // Add to events array
        batchEvents[_batchId].push(newEvent);

        emit EventRecorded(_batchId, _eventType, _location, _actorId);
    }

    /**
     * @dev Record document for a batch
     * @param _batchId Batch ID
     * @param _documentType Type of document
     * @param _ipfsHash IPFS hash of the document
     * @param _issuer ID of the issuer
     */
    function recordDocument(
        string memory _batchId,
        string memory _documentType,
        string memory _ipfsHash,
        string memory _issuer
    ) public whenNotPaused {
        require(bytes(batches[_batchId].batchId).length > 0, "Batch not found");
        require(batches[_batchId].active, "Batch is not active");

        // Create document
        Document memory newDocument = Document({
            batchId: _batchId,
            documentType: _documentType,
            ipfsHash: _ipfsHash,
            issuer: _issuer,
            timestamp: block.timestamp,
            verified: false
        });

        // Add to documents array
        batchDocuments[_batchId].push(newDocument);

        emit DocumentRecorded(_batchId, _documentType, _ipfsHash, _issuer);
    }

    /**
     * @dev Record environment data for a batch
     * @param _batchId Batch ID
     * @param _temperature Temperature scaled by 100
     * @param _ph pH scaled by 100
     * @param _salinity Salinity scaled by 100
     * @param _dissolvedOxygen Dissolved oxygen scaled by 100
     * @param _metadataHash IPFS hash for other parameters
     */
    function recordEnvironmentData(
        string memory _batchId,
        int256 _temperature,
        int256 _ph,
        int256 _salinity,
        int256 _dissolvedOxygen,
        string memory _metadataHash
    ) public whenNotPaused {
        require(bytes(batches[_batchId].batchId).length > 0, "Batch not found");
        require(batches[_batchId].active, "Batch is not active");

        // Create environment data
        EnvironmentData memory newData = EnvironmentData({
            batchId: _batchId,
            timestamp: block.timestamp,
            temperature: _temperature,
            ph: _ph,
            salinity: _salinity,
            dissolvedOxygen: _dissolvedOxygen,
            metadataHash: _metadataHash
        });

        // Add to environment data array
        batchEnvironmentData[_batchId].push(newData);

        emit EnvironmentDataRecorded(
            _batchId,
            _temperature,
            _ph,
            _salinity,
            _dissolvedOxygen
        );
    }

    // ========== CROSS-CHAIN INTEROPERABILITY ==========

    /**
     * @dev Register cross-chain reference for a batch
     * @param _sourceBatchId Source batch ID on this chain
     * @param _targetChain Target blockchain identifier
     * @param _targetTxHash Transaction hash on the target chain
     * @param _targetBatchId Batch ID on the target chain
     * @param _dataStandard Data standard used for conversion (e.g., "GS1-EPCIS")
     */
    function registerCrossChainReference(
        string memory _sourceBatchId,
        string memory _targetChain,
        string memory _targetTxHash,
        string memory _targetBatchId,
        string memory _dataStandard
    ) public whenNotPaused onlyRole(RELAY_ROLE) {
        require(
            bytes(batches[_sourceBatchId].batchId).length > 0,
            "Source batch not found"
        );

        // Create cross-chain reference
        CrossChainReference memory newReference = CrossChainReference({
            sourceBatchId: _sourceBatchId,
            targetChain: _targetChain,
            targetTxHash: _targetTxHash,
            targetBatchId: _targetBatchId,
            dataStandard: _dataStandard,
            timestamp: block.timestamp,
            verified: false
        });

        // Add to cross-chain references array
        crossChainReferences[_sourceBatchId].push(newReference);

        emit CrossChainReferenceCreated(
            _sourceBatchId,
            _targetChain,
            _targetTxHash
        );
    }

    /**
     * @dev Verify cross-chain reference
     * @param _sourceBatchId Source batch ID
     * @param _targetChain Target blockchain
     * @param _targetTxHash Transaction hash on target chain
     */
    function verifyCrossChainReference(
        string memory _sourceBatchId,
        string memory _targetChain,
        string memory _targetTxHash
    ) public whenNotPaused onlyRole(ADMIN_ROLE) {
        require(
            bytes(batches[_sourceBatchId].batchId).length > 0,
            "Source batch not found"
        );

        bool found = false;
        for (uint i = 0; i < crossChainReferences[_sourceBatchId].length; i++) {
            CrossChainReference storage ref = crossChainReferences[
                _sourceBatchId
            ][i];
            if (
                keccak256(abi.encodePacked(ref.targetChain)) ==
                keccak256(abi.encodePacked(_targetChain)) &&
                keccak256(abi.encodePacked(ref.targetTxHash)) ==
                keccak256(abi.encodePacked(_targetTxHash))
            ) {
                ref.verified = true;
                found = true;
                break;
            }
        }

        require(found, "Cross-chain reference not found");
    }

    // ========== DECENTRALIZED IDENTITIES (DIDs) ==========

    /**
     * @dev Create a new DID
     * @param _did Decentralized Identifier
     * @param _controller Controller address
     * @param _metadataHash IPFS hash of the DID metadata
     */
    function createDID(
        string memory _did,
        address _controller,
        string memory _metadataHash
    ) public whenNotPaused {
        require(bytes(_did).length > 0, "DID cannot be empty");
        require(_controller != address(0), "Invalid controller address");
        require(
            didRegistry[_did].controllers.length == 0,
            "DID already exists"
        );

        address[] memory controllers = new address[](1);
        controllers[0] = _controller;

        DID memory newDID = DID({
            id: _did,
            controllers: controllers,
            created: block.timestamp,
            updated: block.timestamp,
            active: true,
            metadataHash: _metadataHash
        });

        didRegistry[_did] = newDID;
        didControllers[_did][_controller] = true;

        emit DIDCreated(_did, _controller);
    }

    /**
     * @dev Add controller to a DID
     * @param _did Decentralized Identifier
     * @param _controller New controller address
     */
    function addDIDController(
        string memory _did,
        address _controller
    ) public whenNotPaused {
        require(bytes(_did).length > 0, "DID cannot be empty");
        require(_controller != address(0), "Invalid controller address");
        require(
            isDIDController(_did, msg.sender),
            "Not authorized to modify DID"
        );
        require(
            !didControllers[_did][_controller],
            "Controller already exists"
        );

        didRegistry[_did].controllers.push(_controller);
        didControllers[_did][_controller] = true;
        didRegistry[_did].updated = block.timestamp;

        emit DIDControllerAdded(_did, _controller);
    }

    /**
     * @dev Remove controller from a DID
     * @param _did Decentralized Identifier
     * @param _controller Controller address to remove
     */
    function removeDIDController(
        string memory _did,
        address _controller
    ) public whenNotPaused {
        require(bytes(_did).length > 0, "DID cannot be empty");
        require(
            isDIDController(_did, msg.sender),
            "Not authorized to modify DID"
        );
        require(didControllers[_did][_controller], "Controller does not exist");
        require(
            didRegistry[_did].controllers.length > 1,
            "Cannot remove the only controller"
        );

        // Find and remove the controller
        address[] storage controllers = didRegistry[_did].controllers;
        for (uint i = 0; i < controllers.length; i++) {
            if (controllers[i] == _controller) {
                controllers[i] = controllers[controllers.length - 1];
                controllers.pop();
                break;
            }
        }

        didControllers[_did][_controller] = false;
        didRegistry[_did].updated = block.timestamp;

        emit DIDControllerRemoved(_did, _controller);
    }

    /**
     * @dev Deactivate a DID
     * @param _did Decentralized Identifier
     */
    function deactivateDID(string memory _did) public whenNotPaused {
        require(bytes(_did).length > 0, "DID cannot be empty");
        require(
            isDIDController(_did, msg.sender),
            "Not authorized to modify DID"
        );
        require(didRegistry[_did].active, "DID already deactivated");

        didRegistry[_did].active = false;
        didRegistry[_did].updated = block.timestamp;

        emit DIDDeactivated(_did);
    }

    /**
     * @dev Reactivate a DID
     * @param _did Decentralized Identifier
     */
    function reactivateDID(string memory _did) public whenNotPaused {
        require(bytes(_did).length > 0, "DID cannot be empty");
        require(
            isDIDController(_did, msg.sender),
            "Not authorized to modify DID"
        );
        require(!didRegistry[_did].active, "DID already active");

        didRegistry[_did].active = true;
        didRegistry[_did].updated = block.timestamp;

        emit DIDReactivated(_did);
    }

    /**
     * @dev Check if an address is a controller of a DID
     * @param _did Decentralized Identifier
     * @param _controller Controller address to check
     * @return bool True if the address is a controller
     */
    function isDIDController(
        string memory _did,
        address _controller
    ) public view returns (bool) {
        return didControllers[_did][_controller];
    }

    // ========== VERIFIABLE CLAIMS ==========

    /**
     * @dev Issue a claim
     * @param _claimId Unique identifier for the claim
     * @param _issuer DID of the issuer
     * @param _subject DID of the subject
     * @param _claimType Type of claim
     * @param _expiryDays Number of days until the claim expires
     * @param _dataHash IPFS hash of the claim data
     */
    function issueClaim(
        string memory _claimId,
        string memory _issuer,
        string memory _subject,
        string memory _claimType,
        uint256 _expiryDays,
        string memory _dataHash
    ) public whenNotPaused {
        require(bytes(_claimId).length > 0, "Claim ID cannot be empty");
        require(bytes(_issuer).length > 0, "Issuer cannot be empty");
        require(bytes(_subject).length > 0, "Subject cannot be empty");
        require(
            isDIDController(_issuer, msg.sender),
            "Not authorized to issue claims for this DID"
        );
        require(
            bytes(claims[_claimId].id).length == 0,
            "Claim ID already exists"
        );

        // Calculate expiry date
        uint256 expiresAt = block.timestamp + (_expiryDays * 1 days);

        // Create claim
        Claim memory newClaim = Claim({
            id: _claimId,
            issuer: _issuer,
            subject: _subject,
            claimType: _claimType,
            issuedAt: block.timestamp,
            expiresAt: expiresAt,
            dataHash: _dataHash,
            revoked: false
        });

        claims[_claimId] = newClaim;
        subjectClaims[_subject].push(_claimId);

        emit ClaimIssued(_claimId, _issuer, _subject, _claimType, expiresAt);
    }

    /**
     * @dev Revoke a claim
     * @param _claimId ID of the claim to revoke
     */
    function revokeClaim(string memory _claimId) public whenNotPaused {
        require(bytes(_claimId).length > 0, "Claim ID cannot be empty");
        require(bytes(claims[_claimId].id).length > 0, "Claim not found");
        require(!claims[_claimId].revoked, "Claim already revoked");
        require(
            isDIDController(claims[_claimId].issuer, msg.sender) ||
                hasRole(ADMIN_ROLE, msg.sender),
            "Not authorized to revoke this claim"
        );

        claims[_claimId].revoked = true;

        emit ClaimRevoked(_claimId, claims[_claimId].issuer);
    }

    /**
     * @dev Verify a claim
     * @param _claimId ID of the claim to verify
     * @return valid Whether the claim is valid
     * @return reason Reason if claim is invalid
     */
    function verifyClaim(
        string memory _claimId
    ) public view returns (bool valid, string memory reason) {
        require(bytes(_claimId).length > 0, "Claim ID cannot be empty");

        Claim memory claim = claims[_claimId];

        if (bytes(claim.id).length == 0) {
            return (false, "Claim not found");
        }

        if (claim.revoked) {
            return (false, "Claim has been revoked");
        }

        if (block.timestamp > claim.expiresAt) {
            return (false, "Claim has expired");
        }

        if (!didRegistry[claim.issuer].active) {
            return (false, "Issuer DID is not active");
        }

        return (true, "");
    }

    /**
     * @dev Get claims for a subject
     * @param _subject Subject DID
     * @return claimIds Array of claim IDs for the subject
     */
    function getClaimsForSubject(
        string memory _subject
    ) public view returns (string[] memory) {
        return subjectClaims[_subject];
    }

    // ========== QUERY FUNCTIONS ==========

    /**
     * @dev Get batch event count
     * @param _batchId Batch ID
     * @return count Number of events for the batch
     */
    function getBatchEventCount(
        string memory _batchId
    ) public view returns (uint256) {
        return batchEvents[_batchId].length;
    }

    /**
     * @dev Get batch document count
     * @param _batchId Batch ID
     * @return count Number of documents for the batch
     */
    function getBatchDocumentCount(
        string memory _batchId
    ) public view returns (uint256) {
        return batchDocuments[_batchId].length;
    }

    /**
     * @dev Get batch environment data count
     * @param _batchId Batch ID
     * @return count Number of environment data records for the batch
     */
    function getBatchEnvironmentDataCount(
        string memory _batchId
    ) public view returns (uint256) {
        return batchEnvironmentData[_batchId].length;
    }

    /**
     * @dev Get batch event
     * @param _batchId Batch ID
     * @param _index Index of the event
     * @return Event The event
     */
    function getBatchEvent(
        string memory _batchId,
        uint256 _index
    ) public view returns (Event memory) {
        require(
            _index < batchEvents[_batchId].length,
            "Event index out of bounds"
        );
        return batchEvents[_batchId][_index];
    }

    /**
     * @dev Get batch document
     * @param _batchId Batch ID
     * @param _index Index of the document
     * @return Document The document
     */
    function getBatchDocument(
        string memory _batchId,
        uint256 _index
    ) public view returns (Document memory) {
        require(
            _index < batchDocuments[_batchId].length,
            "Document index out of bounds"
        );
        return batchDocuments[_batchId][_index];
    }

    /**
     * @dev Get batch environment data
     * @param _batchId Batch ID
     * @param _index Index of the environment data
     * @return EnvironmentData The environment data
     */
    function getBatchEnvironmentData(
        string memory _batchId,
        uint256 _index
    ) public view returns (EnvironmentData memory) {
        require(
            _index < batchEnvironmentData[_batchId].length,
            "Environment data index out of bounds"
        );
        return batchEnvironmentData[_batchId][_index];
    }
}

/**
 * @title ICrossChainConnector
 * @dev Interface for cross-chain connector contracts
 */
interface ICrossChainConnector {
    function sendMessage(
        string memory targetChain,
        bytes memory message
    ) external returns (bytes32);
    function receiveMessage(
        string memory sourceChain,
        bytes memory message,
        bytes memory proof
    ) external returns (bool);
    function verifyMessage(
        string memory sourceChain,
        bytes memory message,
        bytes memory proof
    ) external view returns (bool);
}
