// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title LogisticsTraceability
 * @dev Optimized smart contract for tracking logistics shipments on blockchain
 */
contract LogisticsTraceability {
    // Structure for storing shipment details with optimized storage
    struct Shipment {
        bool exists;
        bytes32 metadataHash; // IPFS or other content-addressed hash instead of full metadata
        address registeredBy;
        uint256 timestamp;
    }

    // Structure for storing event details with optimized storage
    struct ShipmentEvent {
        bytes32 eventTypeHash; // Hashed event type string
        bytes32 metadataHash; // IPFS or other content-addressed hash
        address recordedBy;
        uint256 timestamp;
    }

    // Mapping from shipment ID hash to shipment details (saves gas by using fixed-length keys)
    mapping(bytes32 => Shipment) private shipments;

    // Mapping from shipment ID hash to array of events
    mapping(bytes32 => ShipmentEvent[]) private shipmentEvents;

    // Mapping from shipment ID hash to snapshot indicators
    mapping(bytes32 => uint256) private shipmentSnapshots;

    // State pruning parameters
    uint256 public pruningThreshold = 10000; // Number of blocks after which old events can be pruned
    uint256 public lastPruningBlock;

    // Sharding support variables
    uint8 public constant MAX_SHARDS = 16;
    mapping(uint8 => address) public shardValidators;
    mapping(bytes32 => uint8) public shipmentShard; // Track which shard a shipment belongs to

    // Events for logging
    event ShipmentRegistered(
        bytes32 indexed shipmentIdHash,
        string shipmentId,
        bytes32 metadataHash,
        address registeredBy,
        uint256 timestamp
    );

    event EventRecorded(
        bytes32 indexed shipmentIdHash,
        bytes32 eventTypeHash,
        bytes32 metadataHash,
        address recordedBy,
        uint256 timestamp
    );

    event ShipmentSnapshotCreated(
        bytes32 indexed shipmentIdHash,
        uint256 blockNumber
    );

    // Owner of the contract
    address public owner;

    // Modifier to restrict functions to contract owner
    modifier onlyOwner() {
        require(msg.sender == owner, "Only owner can call this function");
        _;
    }

    constructor() {
        owner = msg.sender;
        lastPruningBlock = block.number;
    }

    /**
     * @dev Assign a shard validator
     * @param shardId Shard identifier (0-15)
     * @param validator Address of the validator
     */
    function assignShardValidator(
        uint8 shardId,
        address validator
    ) public onlyOwner {
        require(shardId < MAX_SHARDS, "Invalid shard ID");
        shardValidators[shardId] = validator;
    }

    /**
     * @dev Set pruning threshold
     * @param threshold New pruning threshold
     */
    function setPruningThreshold(uint256 threshold) public onlyOwner {
        pruningThreshold = threshold;
    }

    /**
     * @dev Calculate shipment ID hash
     * @param shipmentId Original shipment ID string
     * @return Hash of the shipment ID
     */
    function calculateShipmentIdHash(
        string memory shipmentId
    ) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(shipmentId));
    }

    /**
     * @dev Register a new shipment with optimized storage
     * @param shipmentId Unique identifier for the shipment
     * @param metadataHash IPFS hash or other content-addressed hash
     * @return success Whether the operation was successful
     */
    function registerShipment(
        string memory shipmentId,
        bytes32 metadataHash
    ) public returns (bool) {
        // Check if shipment ID is not empty
        require(bytes(shipmentId).length > 0, "Shipment ID cannot be empty");

        // Hash the shipment ID for gas optimization
        bytes32 shipmentIdHash = calculateShipmentIdHash(shipmentId);

        // Check if shipment doesn't already exist
        require(!shipments[shipmentIdHash].exists, "Shipment already exists");

        // Assign to a shard based on hash
        uint8 shardId = uint8(uint256(shipmentIdHash) % MAX_SHARDS);
        shipmentShard[shipmentIdHash] = shardId;

        // Create and store the shipment
        shipments[shipmentIdHash] = Shipment({
            exists: true,
            metadataHash: metadataHash,
            registeredBy: msg.sender,
            timestamp: block.timestamp
        });

        // Emit event
        emit ShipmentRegistered(
            shipmentIdHash,
            shipmentId,
            metadataHash,
            msg.sender,
            block.timestamp
        );

        return true;
    }

    /**
     * @dev Record an event for a shipment with optimized storage
     * @param shipmentId Unique identifier for the shipment
     * @param eventType Type of event (e.g. "PICKUP", "DELIVERY", etc.)
     * @param metadataHash IPFS hash or other content-addressed hash
     * @return success Whether the operation was successful
     */
    function recordEvent(
        string memory shipmentId,
        string memory eventType,
        bytes32 metadataHash
    ) public returns (bool) {
        // Check if shipment ID is not empty
        require(bytes(shipmentId).length > 0, "Shipment ID cannot be empty");
        require(bytes(eventType).length > 0, "Event type cannot be empty");

        // Hash the shipment ID and event type for gas optimization
        bytes32 shipmentIdHash = calculateShipmentIdHash(shipmentId);
        bytes32 eventTypeHash = keccak256(abi.encodePacked(eventType));

        // Check if shipment exists
        require(shipments[shipmentIdHash].exists, "Shipment does not exist");

        // If using sharding, validate appropriate shard
        uint8 shardId = shipmentShard[shipmentIdHash];
        if (shardValidators[shardId] != address(0)) {
            require(
                msg.sender == shardValidators[shardId] || msg.sender == owner,
                "Not authorized for this shard"
            );
        }

        // Create and store the event
        shipmentEvents[shipmentIdHash].push(
            ShipmentEvent({
                eventTypeHash: eventTypeHash,
                metadataHash: metadataHash,
                recordedBy: msg.sender,
                timestamp: block.timestamp
            })
        );

        // Emit event
        emit EventRecorded(
            shipmentIdHash,
            eventTypeHash,
            metadataHash,
            msg.sender,
            block.timestamp
        );

        // Consider creating a snapshot if many events accumulate
        if (shipmentEvents[shipmentIdHash].length % 100 == 0) {
            createShipmentSnapshot(shipmentIdHash);
        }

        return true;
    }

    /**
     * @dev Create a snapshot of shipment state for more efficient retrieval
     * @param shipmentIdHash Hashed shipment ID
     */
    function createShipmentSnapshot(bytes32 shipmentIdHash) internal {
        shipmentSnapshots[shipmentIdHash] = block.number;
        emit ShipmentSnapshotCreated(shipmentIdHash, block.number);
    }

    /**
     * @dev Run storage pruning to optimize blockchain space
     * @param maxItems Maximum number of items to prune in one call
     */
    function runPruning(uint256 maxItems) public {
        require(
            block.number > lastPruningBlock + pruningThreshold,
            "Pruning threshold not reached"
        );

        // Pruning implementation would depend on specific requirements
        // This is a placeholder for the actual implementation

        lastPruningBlock = block.number;
    }

    /**
     * @dev Verify a shipment with optimized lookup
     * @param shipmentId Unique identifier for the shipment
     * @return exists Whether the shipment exists
     * @return metadataHash The shipment metadata hash
     * @return registeredBy Address that registered the shipment
     * @return timestamp When the shipment was registered
     */
    function verifyShipment(
        string memory shipmentId
    )
        public
        view
        returns (
            bool exists,
            bytes32 metadataHash,
            address registeredBy,
            uint256 timestamp
        )
    {
        bytes32 shipmentIdHash = calculateShipmentIdHash(shipmentId);
        Shipment storage shipment = shipments[shipmentIdHash];

        return (
            shipment.exists,
            shipment.metadataHash,
            shipment.registeredBy,
            shipment.timestamp
        );
    }

    /**
     * @dev Get all events for a shipment with optimized retrieval
     * @param shipmentId Unique identifier for the shipment
     * @param startIndex Start index for pagination
     * @param count Maximum number of events to return
     * @return eventTypeHashes Array of event type hashes
     * @return metadataHashes Array of event metadata hashes
     * @return recordedBy Array of addresses that recorded the events
     * @return timestamps Array of event timestamps
     */
    function getEvents(
        string memory shipmentId,
        uint256 startIndex,
        uint256 count
    )
        public
        view
        returns (
            bytes32[] memory eventTypeHashes,
            bytes32[] memory metadataHashes,
            address[] memory recordedBy,
            uint256[] memory timestamps
        )
    {
        bytes32 shipmentIdHash = calculateShipmentIdHash(shipmentId);

        // Get the events for the shipment
        ShipmentEvent[] storage events = shipmentEvents[shipmentIdHash];
        uint256 eventCount = events.length;

        // Ensure startIndex is valid
        if (startIndex >= eventCount) {
            // Return empty arrays if startIndex is out of bounds
            eventTypeHashes = new bytes32[](0);
            metadataHashes = new bytes32[](0);
            recordedBy = new address[](0);
            timestamps = new uint256[](0);
            return (eventTypeHashes, metadataHashes, recordedBy, timestamps);
        }

        // Calculate how many events to return
        uint256 returnCount = (count > 0 && count < (eventCount - startIndex))
            ? count
            : (eventCount - startIndex);

        // Initialize arrays with the correct size
        eventTypeHashes = new bytes32[](returnCount);
        metadataHashes = new bytes32[](returnCount);
        recordedBy = new address[](returnCount);
        timestamps = new uint256[](returnCount);

        // Populate arrays
        for (uint256 i = 0; i < returnCount; i++) {
            uint256 index = startIndex + i;
            eventTypeHashes[i] = events[index].eventTypeHash;
            metadataHashes[i] = events[index].metadataHash;
            recordedBy[i] = events[index].recordedBy;
            timestamps[i] = events[index].timestamp;
        }

        return (eventTypeHashes, metadataHashes, recordedBy, timestamps);
    }

    /**
     * @dev Get the latest snapshot block number for a shipment
     * @param shipmentId Unique identifier for the shipment
     * @return snapshotBlock The latest snapshot block number
     */
    function getShipmentSnapshot(
        string memory shipmentId
    ) public view returns (uint256 snapshotBlock) {
        bytes32 shipmentIdHash = calculateShipmentIdHash(shipmentId);
        return shipmentSnapshots[shipmentIdHash];
    }
}
