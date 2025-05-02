const { ethers } = require("ethers");
const config = require("../config");
const logger = require("../utils/logger");

const LogisticsContractABI = [
  "event ShipmentRegistered(string shipmentId, string metadata, address registeredBy, uint256 timestamp)",
  "event EventRecorded(string shipmentId, string eventType, string metadata, address recordedBy, uint256 timestamp)",
  "function registerShipment(string memory shipmentId, string memory metadata) public returns (bool)",
  "function recordEvent(string memory shipmentId, string memory eventType, string memory metadata) public returns (bool)",
  "function verifyShipment(string memory shipmentId) public view returns (bool exists, string memory metadata, address registeredBy, uint256 timestamp)",
  "function getEvents(string memory shipmentId) public view returns (string[] memory eventTypes, string[] memory metadataList, address[] memory recordedBy, uint256[] memory timestamps)",
];

class EthereumBlockchainService {
  constructor() {
    this.initialized = false;
    this.provider = null;
    this.wallet = null;
    this.contract = null;
    this.consensusType = config.BLOCKCHAIN.ETHEREUM.CONSENSUS_TYPE || "POS";
    this.validators = new Set();
    this.validatorThreshold =
      config.BLOCKCHAIN.ETHEREUM.VALIDATOR_THRESHOLD || 0.66;

    if (config.BLOCKCHAIN.ETHEREUM.ENABLED) {
      this.initialize();
    }
  }

  initialize() {
    try {
      this.provider = new ethers.providers.JsonRpcProvider(
        config.BLOCKCHAIN.ETHEREUM.NODE_URL
      );

      if (config.BLOCKCHAIN.ETHEREUM.PRIVATE_KEY) {
        this.wallet = new ethers.Wallet(
          config.BLOCKCHAIN.ETHEREUM.PRIVATE_KEY,
          this.provider
        );

        if (config.BLOCKCHAIN.ETHEREUM.CONTRACT_ADDRESS) {
          this.contract = new ethers.Contract(
            config.BLOCKCHAIN.ETHEREUM.CONTRACT_ADDRESS,
            LogisticsContractABI,
            this.wallet
          );

          this.initialized = true;
          logger.info("Ethereum Blockchain Service initialized successfully");
        } else {
          logger.warn(
            "Ethereum Contract Address not provided. Contract interactions will not work."
          );
        }
      } else {
        logger.warn(
          "Ethereum Private Key not provided. Read-only mode activated."
        );

        if (config.BLOCKCHAIN.ETHEREUM.CONTRACT_ADDRESS) {
          this.contract = new ethers.Contract(
            config.BLOCKCHAIN.ETHEREUM.CONTRACT_ADDRESS,
            LogisticsContractABI,
            this.provider
          );

          this.initialized = true;
          logger.info(
            "Ethereum Blockchain Service initialized in read-only mode"
          );
        }
      }
    } catch (error) {
      logger.error(
        `Ethereum Blockchain Service initialization failed: ${error.message}`
      );
      this.initialized = false;
    }
  }

  isInitialized() {
    return this.initialized;
  }

  async registerShipment(shipmentId, metadata) {
    try {
      if (!this.initialized || !this.contract) {
        return {
          success: false,
          txHash: null,
          error: "Blockchain service not initialized",
        };
      }

      const tx = await this.contract.registerShipment(shipmentId, metadata);
      const receipt = await tx.wait();

      logger.info(
        `Shipment registered on blockchain: ${shipmentId}, txHash: ${receipt.transactionHash}`
      );

      return {
        success: true,
        txHash: receipt.transactionHash,
        error: null,
      };
    } catch (error) {
      logger.error(
        `Failed to register shipment on blockchain: ${error.message}`
      );

      return {
        success: false,
        txHash: null,
        error: error.message,
      };
    }
  }

  async recordEvent(shipmentId, eventType, metadata) {
    try {
      if (!this.initialized || !this.contract) {
        return {
          success: false,
          txHash: null,
          error: "Blockchain service not initialized",
        };
      }

      const tx = await this.contract.recordEvent(
        shipmentId,
        eventType,
        metadata
      );
      const receipt = await tx.wait();

      logger.info(
        `Event recorded on blockchain: ${shipmentId}, type: ${eventType}, txHash: ${receipt.transactionHash}`
      );

      return {
        success: true,
        txHash: receipt.transactionHash,
        error: null,
      };
    } catch (error) {
      logger.error(`Failed to record event on blockchain: ${error.message}`);

      return {
        success: false,
        txHash: null,
        error: error.message,
      };
    }
  }

  async verifyShipment(shipmentId) {
    try {
      if (!this.initialized || !this.contract) {
        return {
          exists: false,
          metadata: null,
          registeredBy: null,
          timestamp: null,
          error: "Blockchain service not initialized",
        };
      }

      const result = await this.contract.verifyShipment(shipmentId);

      return {
        exists: result.exists,
        metadata: result.metadata,
        registeredBy: result.registeredBy,
        timestamp: result.timestamp.toNumber(),
        error: null,
      };
    } catch (error) {
      logger.error(`Failed to verify shipment on blockchain: ${error.message}`);

      return {
        exists: false,
        metadata: null,
        registeredBy: null,
        timestamp: null,
        error: error.message,
      };
    }
  }

  async getEvents(shipmentId) {
    try {
      if (!this.initialized || !this.contract) {
        return {
          success: false,
          events: [],
          error: "Blockchain service not initialized",
        };
      }

      const result = await this.contract.getEvents(shipmentId);

      const events = [];
      for (let i = 0; i < result.eventTypes.length; i++) {
        events.push({
          type: result.eventTypes[i],
          metadata: result.metadataList[i],
          recordedBy: result.recordedBy[i],
          timestamp: result.timestamps[i].toNumber(),
        });
      }

      return {
        success: true,
        events,
        error: null,
      };
    } catch (error) {
      logger.error(
        `Failed to get shipment events from blockchain: ${error.message}`
      );

      return {
        success: false,
        events: [],
        error: error.message,
      };
    }
  }
}

module.exports = new EthereumBlockchainService();
