// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title ShipmentRegistry
 * @dev Stores and tracks shipments in a supply chain
 */
contract ShipmentRegistry {
    struct Shipment {
        string trackingNumber;
        string dataHash;
        uint256 timestamp;
        string metadata;
        address registrar;
    }

    // Mapping from shipmentId to Shipment
    mapping(string => Shipment) public shipments;

    // Events
    event ShipmentRegistered(
        string indexed shipmentId,
        string trackingNumber,
        string dataHash
    );

    /**
     * @dev Contract constructor
     */
    constructor() {
        // Initialize contract
    }

    /**
     * @dev Register a new shipment
     * @param shipmentId Unique identifier for the shipment
     * @param trackingNumber Shipment tracking number
     * @param dataHash Hash of the shipment data
     * @param metadata Additional metadata (JSON string)
     */
    function registerShipment(
        string memory shipmentId,
        string memory trackingNumber,
        string memory dataHash,
        string memory metadata
    ) public {
        // Ensure shipment doesn't already exist
        require(
            shipments[shipmentId].timestamp == 0,
            "Shipment already exists"
        );

        // Create new shipment
        shipments[shipmentId] = Shipment({
            trackingNumber: trackingNumber,
            dataHash: dataHash,
            timestamp: block.timestamp,
            metadata: metadata,
            registrar: msg.sender
        });

        // Emit event
        emit ShipmentRegistered(shipmentId, trackingNumber, dataHash);
    }

    /**
     * @dev Get shipment details
     * @param shipmentId Shipment identifier
     * @return trackingNumber The shipment tracking number
     * @return dataHash Hash of the shipment data
     * @return timestamp When the shipment was registered
     * @return metadata Additional metadata
     */
    function getShipment(
        string memory shipmentId
    )
        public
        view
        returns (
            string memory trackingNumber,
            string memory dataHash,
            uint256 timestamp,
            string memory metadata
        )
    {
        // Ensure shipment exists
        require(shipments[shipmentId].timestamp > 0, "Shipment does not exist");

        Shipment storage shipment = shipments[shipmentId];
        return (
            shipment.trackingNumber,
            shipment.dataHash,
            shipment.timestamp,
            shipment.metadata
        );
    }
}
