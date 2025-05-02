const config = require("../config");
const logger = require("../utils/logger");
const ethService = require("../blockchain/ethereum.service");
const storageService = require("./storage.service");
const crypto = require("crypto");

class BridgeService {
  constructor() {
    this.initialized = false;
    this.supportedChains = [];
    this.bridgeEnabled = config.BLOCKCHAIN.BRIDGE.ENABLED || false;
    this.verificationThreshold =
      config.BLOCKCHAIN.BRIDGE.VERIFICATION_THRESHOLD || 0.66;
    this.pendingTransfers = new Map();
    this.confirmedTransfers = new Map();

    if (this.bridgeEnabled) {
      this.initialize();
    }
  }

  initialize() {
    try {
      logger.info("Initializing blockchain bridge service");

      if (config.BLOCKCHAIN.ETHEREUM.ENABLED) {
        this.supportedChains.push({
          id: "ethereum",
          type: "evm",
          config: config.BLOCKCHAIN.ETHEREUM,
        });
      }

      if (config.BLOCKCHAIN.POLYGON.ENABLED) {
        this.supportedChains.push({
          id: "polygon",
          type: "evm",
          config: config.BLOCKCHAIN.POLYGON,
        });
      }

      if (config.BLOCKCHAIN.SUBSTRATE.ENABLED) {
        this.supportedChains.push({
          id: "substrate",
          type: "substrate",
          config: config.BLOCKCHAIN.SUBSTRATE,
        });
      }

      this._loadPendingTransfers();

      this.initialized = true;
      logger.info(
        `Bridge service initialized with ${this.supportedChains.length} supported chains`
      );
    } catch (error) {
      logger.error(`Bridge service initialization failed: ${error.message}`);
      this.initialized = false;
    }
  }

  _loadPendingTransfers() {
    try {
      const pendingTransfers = storageService.readData(
        "bridge",
        "pending_transfers.json"
      );
      if (pendingTransfers) {
        this.pendingTransfers = new Map(Object.entries(pendingTransfers));
        logger.info(
          `Loaded ${this.pendingTransfers.size} pending bridge transfers`
        );
      }

      const confirmedTransfers = storageService.readData(
        "bridge",
        "confirmed_transfers.json"
      );
      if (confirmedTransfers) {
        this.confirmedTransfers = new Map(Object.entries(confirmedTransfers));
        logger.info(
          `Loaded ${this.confirmedTransfers.size} confirmed bridge transfers`
        );
      }
    } catch (error) {
      logger.error(`Failed to load pending transfers: ${error.message}`);
    }
  }

  _savePendingTransfers() {
    try {
      storageService.storeData(
        "bridge",
        "pending_transfers.json",
        Object.fromEntries(this.pendingTransfers)
      );

      storageService.storeData(
        "bridge",
        "confirmed_transfers.json",
        Object.fromEntries(this.confirmedTransfers)
      );
    } catch (error) {
      logger.error(`Failed to save pending transfers: ${error.message}`);
    }
  }

  _createTransferId(shipmentId, sourceChain, targetChain) {
    const timestamp = Date.now();
    const hash = crypto
      .createHash("sha256")
      .update(`${shipmentId}-${sourceChain}-${targetChain}-${timestamp}`)
      .digest("hex");
    return hash.substring(0, 16);
  }

  async initiateTransfer(shipmentId, sourceChain, targetChain) {
    try {
      if (!this.bridgeEnabled) {
        return {
          success: false,
          transferId: null,
          error: "Bridge service is not enabled",
        };
      }

      const sourceChainConfig = this.supportedChains.find(
        (chain) => chain.id === sourceChain
      );
      const targetChainConfig = this.supportedChains.find(
        (chain) => chain.id === targetChain
      );

      if (!sourceChainConfig) {
        return {
          success: false,
          transferId: null,
          error: `Source chain ${sourceChain} is not supported`,
        };
      }

      if (!targetChainConfig) {
        return {
          success: false,
          transferId: null,
          error: `Target chain ${targetChain} is not supported`,
        };
      }

      const verificationResult = await this._verifyShipmentOnChain(
        shipmentId,
        sourceChain
      );
      if (!verificationResult.exists) {
        return {
          success: false,
          transferId: null,
          error: `Shipment ${shipmentId} not found on ${sourceChain}`,
        };
      }

      const transferId = this._createTransferId(
        shipmentId,
        sourceChain,
        targetChain
      );

      this.pendingTransfers.set(transferId, {
        shipmentId,
        sourceChain,
        targetChain,
        status: "pending",
        initiatedAt: Date.now(),
        metadata: verificationResult.metadata,
        validations: [],
        sourceTxHash: null,
        targetTxHash: null,
      });

      this._savePendingTransfers();

      logger.info(
        `Initiated cross-chain transfer: ${transferId} for shipment ${shipmentId} from ${sourceChain} to ${targetChain}`
      );

      this._processTransfer(transferId).catch((error) => {
        logger.error(
          `Error processing transfer ${transferId}: ${error.message}`
        );
      });

      return {
        success: true,
        transferId,
        error: null,
      };
    } catch (error) {
      logger.error(`Failed to initiate cross-chain transfer: ${error.message}`);

      return {
        success: false,
        transferId: null,
        error: error.message,
      };
    }
  }

  async _processTransfer(transferId) {
    try {
      const transfer = this.pendingTransfers.get(transferId);
      if (!transfer) {
        logger.error(`Transfer ${transferId} not found`);
        return;
      }

      transfer.status = "processing";
      this._savePendingTransfers();

      const result = await this._createShipmentOnTargetChain(
        transfer.shipmentId,
        transfer.metadata,
        transfer.targetChain
      );

      if (result.success) {
        transfer.status = "completed";
        transfer.targetTxHash = result.txHash;
        transfer.completedAt = Date.now();

        this.confirmedTransfers.set(transferId, transfer);
        this.pendingTransfers.delete(transferId);

        logger.info(
          `Completed cross-chain transfer: ${transferId} for shipment ${transfer.shipmentId}`
        );
      } else {
        transfer.status = "failed";
        transfer.error = result.error;

        logger.error(
          `Failed cross-chain transfer: ${transferId} - ${result.error}`
        );
      }

      this._savePendingTransfers();
    } catch (error) {
      logger.error(`Error processing transfer ${transferId}: ${error.message}`);

      const transfer = this.pendingTransfers.get(transferId);
      if (transfer) {
        transfer.status = "failed";
        transfer.error = error.message;
        this._savePendingTransfers();
      }
    }
  }

  async _verifyShipmentOnChain(shipmentId, chainId) {
    try {
      if (chainId === "ethereum" && ethService.isInitialized()) {
        const verification = await ethService.verifyShipment(shipmentId);

        return {
          exists: verification.exists,
          metadata: verification.metadata,
          error: verification.error,
        };
      }

      return {
        exists: false,
        metadata: null,
        error: `Verification for chain ${chainId} not implemented`,
      };
    } catch (error) {
      logger.error(
        `Verification error for ${shipmentId} on ${chainId}: ${error.message}`
      );

      return {
        exists: false,
        metadata: null,
        error: error.message,
      };
    }
  }

  async _createShipmentOnTargetChain(shipmentId, metadata, chainId) {
    try {
      if (chainId === "ethereum" && ethService.isInitialized()) {
        return await ethService.registerShipment(shipmentId, metadata);
      }

      return {
        success: false,
        txHash: null,
        error: `Registration for chain ${chainId} not implemented`,
      };
    } catch (error) {
      logger.error(
        `Registration error for ${shipmentId} on ${chainId}: ${error.message}`
      );

      return {
        success: false,
        txHash: null,
        error: error.message,
      };
    }
  }

  getTransferStatus(transferId) {
    if (this.pendingTransfers.has(transferId)) {
      return this.pendingTransfers.get(transferId);
    }

    if (this.confirmedTransfers.has(transferId)) {
      return this.confirmedTransfers.get(transferId);
    }

    return null;
  }

  getSupportedChains() {
    return this.supportedChains.map((chain) => chain.id);
  }

  isInitialized() {
    return this.initialized;
  }
}

module.exports = new BridgeService();
