// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title SensorDataLog
 * @dev Stores IoT sensor data related to shipments
 */
contract SensorDataLog {
    struct SensorReading {
        string shipmentId;
        string readingType; // temperature, humidity, shock, etc.
        int256 value;
        uint256 timestamp;
        string location;
        string metadata;
        address logger;
    }

    // Mapping from readingId to SensorReading
    mapping(string => SensorReading) public readings;

    // Track readings by shipment
    mapping(string => string[]) private shipmentReadings;

    // Events
    event SensorDataLogged(
        string indexed shipmentId,
        string indexed readingId,
        string readingType,
        int256 value
    );

    event AlertTriggered(
        string indexed shipmentId,
        string indexed readingId,
        string readingType,
        int256 value,
        string alertType
    );

    /**
     * @dev Contract constructor
     */
    constructor() {
        // Initialize contract
    }

    /**
     * @dev Log a sensor reading
     * @param shipmentId Shipment identifier
     * @param readingId Unique reading identifier
     * @param readingType Type of reading (temperature, humidity, etc.)
     * @param value Sensor value (can be negative, scaled by 100 for decimals)
     * @param location Location string
     * @param metadata Additional metadata (JSON string)
     * @param minThreshold Minimum threshold for alert
     * @param maxThreshold Maximum threshold for alert
     */
    function logSensorReading(
        string memory shipmentId,
        string memory readingId,
        string memory readingType,
        int256 value,
        string memory location,
        string memory metadata,
        int256 minThreshold,
        int256 maxThreshold
    ) public {
        // Ensure reading doesn't already exist
        require(readings[readingId].timestamp == 0, "Reading already exists");

        // Create new reading
        readings[readingId] = SensorReading({
            shipmentId: shipmentId,
            readingType: readingType,
            value: value,
            timestamp: block.timestamp,
            location: location,
            metadata: metadata,
            logger: msg.sender
        });

        // Add to shipment readings
        shipmentReadings[shipmentId].push(readingId);

        // Emit event
        emit SensorDataLogged(shipmentId, readingId, readingType, value);

        // Check thresholds and trigger alerts if needed
        if (value < minThreshold) {
            emit AlertTriggered(
                shipmentId,
                readingId,
                readingType,
                value,
                "BELOW_THRESHOLD"
            );
        } else if (value > maxThreshold) {
            emit AlertTriggered(
                shipmentId,
                readingId,
                readingType,
                value,
                "ABOVE_THRESHOLD"
            );
        }
    }

    /**
     * @dev Get sensor reading details
     * @param readingId Reading identifier
     * @return shipmentId The associated shipment ID
     * @return readingType The type of reading
     * @return value The sensor value
     * @return timestamp When the reading was logged
     * @return location The location string
     * @return metadata Additional metadata
     */
    function getSensorReading(
        string memory readingId
    )
        public
        view
        returns (
            string memory shipmentId,
            string memory readingType,
            int256 value,
            uint256 timestamp,
            string memory location,
            string memory metadata
        )
    {
        // Ensure reading exists
        require(readings[readingId].timestamp > 0, "Reading does not exist");

        SensorReading storage reading = readings[readingId];
        return (
            reading.shipmentId,
            reading.readingType,
            reading.value,
            reading.timestamp,
            reading.location,
            reading.metadata
        );
    }

    /**
     * @dev Get all readings for a shipment
     * @param shipmentId Shipment identifier
     * @return readingIds Array of reading IDs
     */
    function getShipmentReadings(
        string memory shipmentId
    ) public view returns (string[] memory readingIds) {
        return shipmentReadings[shipmentId];
    }

    /**
     * @dev Get the latest reading of a specific type for a shipment
     * @param shipmentId Shipment identifier
     * @param readingType Type of reading to find
     * @return readingId The reading ID
     * @return value The sensor value
     * @return timestamp When the reading was logged
     */
    function getLatestReading(
        string memory shipmentId,
        string memory readingType
    )
        public
        view
        returns (string memory readingId, int256 value, uint256 timestamp)
    {
        string[] memory readingIds = shipmentReadings[shipmentId];

        uint256 latestTime = 0;
        string memory latestId = "";
        int256 latestValue = 0;

        for (uint i = 0; i < readingIds.length; i++) {
            SensorReading storage reading = readings[readingIds[i]];

            if (
                keccak256(bytes(reading.readingType)) ==
                keccak256(bytes(readingType)) &&
                reading.timestamp > latestTime
            ) {
                latestTime = reading.timestamp;
                latestId = readingIds[i];
                latestValue = reading.value;
            }
        }

        require(latestTime > 0, "No readings found");

        return (latestId, latestValue, latestTime);
    }
}
