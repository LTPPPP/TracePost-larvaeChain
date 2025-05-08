// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/Pausable.sol";
import "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/math/SafeMath.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/token/ERC721/extensions/ERC721Enumerable.sol";

// Add at the top of the file, before the LogisticsTraceability contract
interface ICrossChainConnector {
    function sendMessage(
        string memory destinationChain,
        bytes memory payload
    ) external returns (bytes32);
    function receiveMessage(
        string memory sourceChain,
        bytes memory payload,
        bytes memory proof
    ) external returns (bool);
    function verifyMessage(
        string memory sourceChain,
        bytes32 messageId,
        bytes memory proof
    ) external view returns (bool);
    function getChainStatus(
        string memory chainId
    ) external view returns (string memory);
}

/**
 * @title LogisticsTraceability
 * @dev Smart contract for managing logistics traceability with blockchain interoperability, DID support, and NFT capabilities
 * @custom:experimental This is an experimental contract for TracePost-larvaeChain
 */
contract LogisticsTraceability is
    AccessControl,
    Pausable,
    Initializable,
    ERC721URIStorage,
    ERC721Enumerable
{
    using SafeMath for uint256;
    using ECDSA for bytes32;

    // Role definitions
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    bytes32 public constant HATCHERY_ROLE = keccak256("HATCHERY_ROLE");
    bytes32 public constant FARM_ROLE = keccak256("FARM_ROLE");
    bytes32 public constant PROCESSOR_ROLE = keccak256("PROCESSOR_ROLE");
    bytes32 public constant CERTIFIER_ROLE = keccak256("CERTIFIER_ROLE");
    bytes32 public constant RELAY_ROLE = keccak256("RELAY_ROLE"); // For cross-chain operations
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE"); // For NFT minting

    // Cross-chain connector interface for interoperability
    ICrossChainConnector public crossChainConnector;

    // DID Registry
    struct DID {
        string did;
        address owner;
        bool active;
        uint256 created;
        uint256 updated;
    }

    mapping(string => DID) public didRegistry; // did -> DID struct
    mapping(string => string) public didDocuments; // didHash -> IPFS hash of DID document
    mapping(string => mapping(address => bool)) public didControllers; // didHash -> controller -> isController

    // Batch data structure
    struct Batch {
        string batchId;
        string hatcheryId;
        string species;
        uint256 quantity;
        string status;
        uint256 created;
        uint256 updated;
        bool active;
        uint256 tokenId; // NFT token ID if minted
        bool isTokenized; // Whether batch has been tokenized
    }

    // Event data structure
    struct BatchEvent {
        string eventId;
        string batchId;
        string eventType;
        string location;
        uint256 timestamp;
        string metadataHash; // IPFS hash of event metadata
    }

    // Batch registry
    mapping(string => Batch) public batches; // batchId -> Batch
    string[] private batchIds; // Array of all batch IDs

    // Events registry
    mapping(string => BatchEvent[]) public batchEvents; // batchId -> array of events

    // Batch to owner mapping for NFTs
    mapping(uint256 => string) public tokenToBatch; // tokenId -> batchId
    mapping(string => uint256) public batchToToken; // batchId -> tokenId

    // NFT-specific variables
    uint256 private _nextTokenId;
    string private _baseTokenURI;

    // Events
    event BatchCreated(
        string batchId,
        string hatcheryId,
        string species,
        uint256 quantity,
        uint256 timestamp
    );
    event BatchStatusUpdated(string batchId, string status, uint256 timestamp);
    event BatchEventRecorded(
        string batchId,
        string eventId,
        string eventType,
        uint256 timestamp
    );
    event BatchTokenized(
        string batchId,
        uint256 tokenId,
        address owner,
        uint256 timestamp
    );
    event BatchDocumentAdded(
        string batchId,
        string docType,
        string ipfsHash,
        uint256 timestamp
    );

    // Cross-chain events
    event CrossChainBatchShared(
        string batchId,
        string destinationChain,
        bytes32 messageId,
        uint256 timestamp
    );
    event CrossChainBatchReceived(
        string batchId,
        string sourceChain,
        uint256 timestamp
    );

    /**
     * @dev Constructor initializes the contract with a name and symbol for the NFT
     */
    constructor() ERC721("TracePost Batch NFT", "TPBATCH") {
        _setupRole(DEFAULT_ADMIN_ROLE, _msgSender());
        _setupRole(ADMIN_ROLE, _msgSender());
        _setupRole(MINTER_ROLE, _msgSender());
        _nextTokenId = 1;
        _baseTokenURI = "https://trace.viechain.com/api/v1/tokens/";
    }

    /**
     * @dev Initialize method for proxy deployment
     */
    function initialize(
        address admin,
        string memory baseURI
    ) public initializer {
        _setupRole(DEFAULT_ADMIN_ROLE, admin);
        _setupRole(ADMIN_ROLE, admin);
        _setupRole(MINTER_ROLE, admin);
        _nextTokenId = 1;
        _baseTokenURI = baseURI;
    }

    /**
     * @dev Create a new batch
     * @param batchId Unique identifier for the batch
     * @param hatcheryId Identifier for the hatchery
     * @param species Species of shrimp larvae
     * @param quantity Number of larvae in the batch
     */
    function createBatch(
        string memory batchId,
        string memory hatcheryId,
        string memory species,
        uint256 quantity
    ) public whenNotPaused {
        require(
            hasRole(HATCHERY_ROLE, _msgSender()) ||
                hasRole(ADMIN_ROLE, _msgSender()),
            "Must have hatchery or admin role"
        );
        require(
            bytes(batches[batchId].batchId).length == 0,
            "Batch ID already exists"
        );

        batches[batchId] = Batch({
            batchId: batchId,
            hatcheryId: hatcheryId,
            species: species,
            quantity: quantity,
            status: "created",
            created: block.timestamp,
            updated: block.timestamp,
            active: true,
            tokenId: 0,
            isTokenized: false
        });

        batchIds.push(batchId);

        emit BatchCreated(
            batchId,
            hatcheryId,
            species,
            quantity,
            block.timestamp
        );
    }

    /**
     * @dev Update the status of a batch
     * @param batchId Batch identifier
     * @param status New status
     */
    function updateBatchStatus(
        string memory batchId,
        string memory status
    ) public whenNotPaused {
        require(
            hasRole(HATCHERY_ROLE, _msgSender()) ||
                hasRole(FARM_ROLE, _msgSender()) ||
                hasRole(PROCESSOR_ROLE, _msgSender()) ||
                hasRole(ADMIN_ROLE, _msgSender()),
            "Must have appropriate role"
        );
        require(
            bytes(batches[batchId].batchId).length > 0,
            "Batch does not exist"
        );
        require(batches[batchId].active, "Batch is not active");

        batches[batchId].status = status;
        batches[batchId].updated = block.timestamp;

        emit BatchStatusUpdated(batchId, status, block.timestamp);
    }

    /**
     * @dev Record an event for a batch
     * @param batchId Batch identifier
     * @param eventId Unique event identifier
     * @param eventType Type of event
     * @param location Location where event occurred
     * @param metadataHash IPFS hash of event metadata
     */
    function recordBatchEvent(
        string memory batchId,
        string memory eventId,
        string memory eventType,
        string memory location,
        string memory metadataHash
    ) public whenNotPaused {
        require(
            hasRole(HATCHERY_ROLE, _msgSender()) ||
                hasRole(FARM_ROLE, _msgSender()) ||
                hasRole(PROCESSOR_ROLE, _msgSender()) ||
                hasRole(CERTIFIER_ROLE, _msgSender()) ||
                hasRole(ADMIN_ROLE, _msgSender()),
            "Must have appropriate role"
        );
        require(
            bytes(batches[batchId].batchId).length > 0,
            "Batch does not exist"
        );
        require(batches[batchId].active, "Batch is not active");

        BatchEvent memory newEvent = BatchEvent({
            eventId: eventId,
            batchId: batchId,
            eventType: eventType,
            location: location,
            timestamp: block.timestamp,
            metadataHash: metadataHash
        });

        batchEvents[batchId].push(newEvent);

        emit BatchEventRecorded(batchId, eventId, eventType, block.timestamp);
    }

    /**
     * @dev Tokenize a batch by minting an NFT
     * @param batchId Batch identifier
     * @param recipient Address to receive the NFT
     * @param tokenURI URI for token metadata
     * @return tokenId The minted token ID
     */
    function tokenizeBatch(
        string memory batchId,
        address recipient,
        string memory tokenURI
    ) public whenNotPaused returns (uint256) {
        require(
            hasRole(MINTER_ROLE, _msgSender()) ||
                hasRole(ADMIN_ROLE, _msgSender()),
            "Must have minter or admin role"
        );
        require(
            bytes(batches[batchId].batchId).length > 0,
            "Batch does not exist"
        );
        require(!batches[batchId].isTokenized, "Batch is already tokenized");
        require(batches[batchId].active, "Batch is not active");

        uint256 tokenId = _nextTokenId;
        _nextTokenId++;

        _mint(recipient, tokenId);
        _setTokenURI(tokenId, tokenURI);

        // Update batch record
        batches[batchId].tokenId = tokenId;
        batches[batchId].isTokenized = true;
        batches[batchId].updated = block.timestamp;

        // Update mappings
        tokenToBatch[tokenId] = batchId;
        batchToToken[batchId] = tokenId;

        emit BatchTokenized(batchId, tokenId, recipient, block.timestamp);

        return tokenId;
    }

    /**
     * @dev Get batch details
     * @param batchId Batch identifier
     * @return Batch details
     */
    function getBatch(
        string memory batchId
    ) public view returns (Batch memory) {
        require(
            bytes(batches[batchId].batchId).length > 0,
            "Batch does not exist"
        );
        return batches[batchId];
    }

    /**
     * @dev Get batch events
     * @param batchId Batch identifier
     * @return Array of batch events
     */
    function getBatchEvents(
        string memory batchId
    ) public view returns (BatchEvent[] memory) {
        return batchEvents[batchId];
    }

    /**
     * @dev Get total number of batches
     * @return Total number of batches
     */
    function getTotalBatches() public view returns (uint256) {
        return batchIds.length;
    }

    /**
     * @dev Get batch ID by index
     * @param index Index in the batch array
     * @return Batch identifier
     */
    function getBatchIdByIndex(
        uint256 index
    ) public view returns (string memory) {
        require(index < batchIds.length, "Index out of bounds");
        return batchIds[index];
    }

    /**
     * @dev Get batch for a token ID
     * @param tokenId Token identifier
     * @return Batch identifier
     */
    function getBatchByTokenId(
        uint256 tokenId
    ) public view returns (string memory) {
        string memory batchId = tokenToBatch[tokenId];
        require(
            bytes(batchId).length > 0,
            "Token ID not associated with any batch"
        );
        return batchId;
    }

    /**
     * @dev Get token ID for a batch
     * @param batchId Batch identifier
     * @return Token identifier
     */
    function getTokenIdByBatch(
        string memory batchId
    ) public view returns (uint256) {
        require(batches[batchId].isTokenized, "Batch is not tokenized");
        return batchToToken[batchId];
    }

    /**
     * @dev Share batch data with another blockchain
     * @param batchId Batch identifier
     * @param destinationChain Destination chain identifier
     * @return messageId Message identifier in the cross-chain system
     */
    function shareBatchWithChain(
        string memory batchId,
        string memory destinationChain
    ) public whenNotPaused returns (bytes32) {
        require(
            hasRole(RELAY_ROLE, _msgSender()) ||
                hasRole(ADMIN_ROLE, _msgSender()),
            "Must have relay or admin role"
        );
        require(
            bytes(batches[batchId].batchId).length > 0,
            "Batch does not exist"
        );
        require(
            address(crossChainConnector) != address(0),
            "Cross-chain connector not set"
        );

        // Encode batch data for cross-chain transfer
        bytes memory payload = abi.encode(
            batches[batchId].batchId,
            batches[batchId].hatcheryId,
            batches[batchId].species,
            batches[batchId].quantity,
            batches[batchId].status,
            batches[batchId].created,
            batches[batchId].updated
        );

        bytes32 messageId = crossChainConnector.sendMessage(
            destinationChain,
            payload
        );

        emit CrossChainBatchShared(
            batchId,
            destinationChain,
            messageId,
            block.timestamp
        );

        return messageId;
    }

    /**
     * @dev Receive batch data from another blockchain
     * @param sourceChain Source chain identifier
     * @param payload Encoded batch data
     * @param proof Proof of message authenticity
     * @return success Whether the operation was successful
     */
    function receiveBatchFromChain(
        string memory sourceChain,
        bytes memory payload,
        bytes memory proof
    ) public whenNotPaused returns (bool) {
        require(
            hasRole(RELAY_ROLE, _msgSender()) ||
                hasRole(ADMIN_ROLE, _msgSender()),
            "Must have relay or admin role"
        );
        require(
            address(crossChainConnector) != address(0),
            "Cross-chain connector not set"
        );

        bool success = crossChainConnector.receiveMessage(
            sourceChain,
            payload,
            proof
        );
        require(success, "Failed to verify cross-chain message");

        // Decode batch data
        (
            string memory batchId,
            string memory hatcheryId,
            string memory species,
            uint256 quantity,
            string memory status,
            uint256 created,
            uint256 updated
        ) = abi.decode(
                payload,
                (string, string, string, uint256, string, uint256, uint256)
            );

        // Check if batch already exists
        if (bytes(batches[batchId].batchId).length == 0) {
            // Create new batch
            batches[batchId] = Batch({
                batchId: batchId,
                hatcheryId: hatcheryId,
                species: species,
                quantity: quantity,
                status: status,
                created: created,
                updated: updated,
                active: true,
                tokenId: 0,
                isTokenized: false
            });

            batchIds.push(batchId);
        } else {
            // Update existing batch
            batches[batchId].status = status;
            batches[batchId].updated = updated;
        }

        emit CrossChainBatchReceived(batchId, sourceChain, block.timestamp);

        return true;
    }

    /**
     * @dev Set the cross-chain connector
     * @param connector Address of the cross-chain connector
     */
    function setCrossChainConnector(
        address connector
    ) public onlyRole(ADMIN_ROLE) {
        crossChainConnector = ICrossChainConnector(connector);
    }

    /**
     * @dev Set the base token URI
     * @param baseURI Base URI for token metadata
     */
    function setBaseTokenURI(
        string memory baseURI
    ) public onlyRole(ADMIN_ROLE) {
        _baseTokenURI = baseURI;
    }

    /**
     * @dev Get the base token URI
     * @return Base URI for token metadata
     */
    function getBaseTokenURI() public view returns (string memory) {
        return _baseTokenURI;
    }

    /**
     * @dev Pause contract (disables most functions)
     */
    function pause() public onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /**
     * @dev Unpause contract
     */
    function unpause() public onlyRole(ADMIN_ROLE) {
        _unpause();
    }

    /**
     * @dev Hook that is called before any token transfer
     */
    function _beforeTokenTransfer(
        address from,
        address to,
        uint256 firstTokenId,
        uint256 batchSize
    ) internal override(ERC721, ERC721Enumerable) whenNotPaused {
        super._beforeTokenTransfer(from, to, firstTokenId, batchSize);
    }

    /**
     * @dev Burn a token
     * @param tokenId Token identifier
     */
    function burn(uint256 tokenId) public {
        require(
            _isApprovedOrOwner(_msgSender(), tokenId),
            "Caller is not token owner or approved"
        );

        string memory batchId = tokenToBatch[tokenId];
        if (bytes(batchId).length > 0) {
            // Update batch record
            batches[batchId].isTokenized = false;
            batches[batchId].tokenId = 0;

            // Clear mappings
            delete batchToToken[batchId];
            delete tokenToBatch[tokenId];
        }

        _burn(tokenId);
    }

    // Required overrides
    function supportsInterface(
        bytes4 interfaceId
    )
        public
        view
        override(ERC721, ERC721Enumerable, AccessControl)
        returns (bool)
    {
        return super.supportsInterface(interfaceId);
    }

    function tokenURI(
        uint256 tokenId
    ) public view override(ERC721, ERC721URIStorage) returns (string memory) {
        return super.tokenURI(tokenId);
    }

    function _burn(
        uint256 tokenId
    ) internal override(ERC721, ERC721URIStorage) {
        super._burn(tokenId);
    }
}
