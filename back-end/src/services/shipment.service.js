/**
 * Shipment service for managing logistics shipments
 * Integrates with blockchain for traceability
 */

const { v4: uuidv4 } = require("uuid");
const ethService = require("../blockchain/ethereum.service");
const shardingService = require("./sharding.service");
const bridgeService = require("./bridge.service");
const storageService = require("./storage.service");
const logger = require("../utils/logger");
const config = require("../config");

class ShipmentService {
  constructor() {
    this.useBlockchain = config.BLOCKCHAIN.ENABLED;
    this.useFastCache = config.SHIPMENT.FAST_CACHE_ENABLED || true;
    this.useSharding = config.BLOCKCHAIN.SHARDING.ENABLED || false;
    this.useBridge = config.BLOCKCHAIN.BRIDGE.ENABLED || false;
    this.blockchainWriteInterval =
      config.SHIPMENT.BLOCKCHAIN_WRITE_INTERVAL || 0;
    this.pendingEvents = new Map(); // For batch processing of events
    this.lastBatchTime = Date.now();

    // Initialize services if enabled
    this._initializeServices();

    // Set up batch processing interval if configured
    if (this.blockchainWriteInterval > 0) {
      setInterval(
        () => this._processPendingEvents(),
        this.blockchainWriteInterval
      );
      logger.info(
        `Set up batch processing interval: ${this.blockchainWriteInterval}ms`
      );
    }
  }

  /**
   * Initialize required services
   * @private
   */
  _initializeServices() {
    // Verify blockchain service is available if enabled
    if (this.useBlockchain) {
      if (!ethService.isInitialized()) {
        logger.warn(
          "Ethereum service not initialized, blockchain features will be disabled"
        );
        this.useBlockchain = false;
      }
    }

    // Verify sharding service is available if enabled
    if (this.useSharding) {
      if (!shardingService.isInitialized()) {
        logger.warn(
          "Sharding service not initialized, sharding features will be disabled"
        );
        this.useSharding = false;
      }
    }

    // Verify bridge service is available if enabled
    if (this.useBridge) {
      if (!bridgeService.isInitialized()) {
        logger.warn(
          "Bridge service not initialized, cross-chain features will be disabled"
        );
        this.useBridge = false;
      }
    }
  }

  /**
   * Create a new shipment
   * @param {Object} shipmentData - The shipment data
   * @param {boolean} recordOnBlockchain - Whether to record on blockchain
   * @returns {Promise<Object>} The created shipment
   */
  async createShipment(shipmentData, recordOnBlockchain = true) {
    try {
      // Generate a unique ID if not provided
      const shipmentId = shipmentData.id || uuidv4();
      const shipment = {
        id: shipmentId,
        ...shipmentData,
        createdAt: Date.now(),
        events: [],
        blockchainRecorded: false,
        blockchainTxHash: null,
      };

      // Determine target blockchain if sharding is enabled
      let targetChain = "ethereum"; // Default chain
      if (this.useSharding) {
        const shardId = shardingService.getShipmentShard(shipmentId);
        shipment.shardId = shardId;

        // If we're not responsible for this shard, determine the appropriate chain
        if (!shardingService.isResponsibleForShard(shardId)) {
          if (this.useBridge && bridgeService.getSupportedChains().length > 1) {
            // Use another chain if bridge is available and we have multiple chains
            const chains = bridgeService.getSupportedChains();
            targetChain = chains.find((c) => c !== "ethereum") || "ethereum";
            shipment.targetChain = targetChain;
          }
        }
      }

      // Store shipment data locally
      storageService.storeData("shipments", `${shipmentId}.json`, shipment);

      // Record on blockchain if enabled
      if (this.useBlockchain && recordOnBlockchain) {
        // Prepare metadata (JSON string or IPFS hash)
        const metadata = JSON.stringify({
          description: shipment.description || "",
          origin: shipment.origin || "",
          destination: shipment.destination || "",
          type: shipment.type || "",
          timestamp: shipment.createdAt,
        });

        // Convert metadata to bytes32 hash if using optimized contract
        let metadataHash = metadata;
        if (config.BLOCKCHAIN.USE_OPTIMIZED_CONTRACT) {
          // Simple hash function, could use proper IPFS hash in production
          const crypto = require("crypto");
          metadataHash =
            "0x" + crypto.createHash("sha256").update(metadata).digest("hex");
        }

        // Determine if we should record now or batch for later
        if (this.blockchainWriteInterval > 0) {
          // Add to pending events for batch processing
          if (!this.pendingEvents.has(shipmentId)) {
            this.pendingEvents.set(shipmentId, []);
          }

          this.pendingEvents.get(shipmentId).push({
            type: "registerShipment",
            shipmentId,
            metadataHash,
            chain: targetChain,
          });

          logger.info(
            `Queued shipment ${shipmentId} for blockchain recording in next batch`
          );
        } else {
          // Record immediately
          const result = await this._recordShipmentOnBlockchain(
            shipmentId,
            metadataHash,
            targetChain
          );

          if (result.success) {
            shipment.blockchainRecorded = true;
            shipment.blockchainTxHash = result.txHash;

            // Update stored shipment with blockchain info
            storageService.storeData(
              "shipments",
              `${shipmentId}.json`,
              shipment
            );
          }
        }
      }

      return shipment;
    } catch (error) {
      logger.error(`Failed to create shipment: ${error.message}`);
      throw error;
    }
  }

  /**
   * Record a shipment event
   * @param {string} shipmentId - The shipment ID
   * @param {Object} eventData - The event data
   * @param {boolean} recordOnBlockchain - Whether to record on blockchain
   * @returns {Promise<Object>} The recorded event
   */
  async recordEvent(shipmentId, eventData, recordOnBlockchain = true) {
    try {
      // Load the shipment
      const shipment = storageService.readData(
        "shipments",
        `${shipmentId}.json`
      );
      if (!shipment) {
        throw new Error(`Shipment ${shipmentId} not found`);
      }

      // Create the event
      const event = {
        id: uuidv4(),
        shipmentId,
        type: eventData.type,
        location: eventData.location,
        timestamp: Date.now(),
        data: eventData.data || {},
        blockchainRecorded: false,
        blockchainTxHash: null,
        ...eventData,
      };

      // Add to shipment events
      shipment.events.push(event);

      // Find target chain from shipment if set, or default to ethereum
      const targetChain = shipment.targetChain || "ethereum";

      // Update the shipment in storage
      storageService.storeData("shipments", `${shipmentId}.json`, shipment);

      // Also store event separately for easy access
      storageService.storeData("events", `${event.id}.json`, event);

      // Get shipment shard if sharding is enabled
      let shardId = null;
      if (this.useSharding) {
        shardId =
          shipment.shardId || shardingService.getShipmentShard(shipmentId);
      }

      // Record on blockchain if enabled
      if (this.useBlockchain && recordOnBlockchain) {
        // Check if we're responsible for this shard if sharding is enabled
        const shouldRecord =
          !this.useSharding || shardingService.isResponsibleForShard(shardId);

        if (shouldRecord) {
          // Prepare metadata
          const metadata = JSON.stringify({
            type: event.type,
            location: event.location,
            timestamp: event.timestamp,
            data: event.data,
          });

          // Convert metadata to bytes32 hash if using optimized contract
          let metadataHash = metadata;
          if (config.BLOCKCHAIN.USE_OPTIMIZED_CONTRACT) {
            // Simple hash function, could use proper IPFS hash in production
            const crypto = require("crypto");
            metadataHash =
              "0x" + crypto.createHash("sha256").update(metadata).digest("hex");
          }

          // Determine if we should record now or batch for later
          if (this.blockchainWriteInterval > 0) {
            // Add to pending events for batch processing
            if (!this.pendingEvents.has(shipmentId)) {
              this.pendingEvents.set(shipmentId, []);
            }

            this.pendingEvents.get(shipmentId).push({
              type: "recordEvent",
              shipmentId,
              eventType: event.type,
              metadataHash,
              chain: targetChain,
            });

            logger.info(
              `Queued event ${event.id} for blockchain recording in next batch`
            );
          } else {
            // Record immediately
            const result = await this._recordEventOnBlockchain(
              shipmentId,
              event.type,
              metadataHash,
              targetChain
            );

            if (result.success) {
              event.blockchainRecorded = true;
              event.blockchainTxHash = result.txHash;

              // Update stored event with blockchain info
              storageService.storeData("events", `${event.id}.json`, event);

              // Also update in shipment
              const shipmentForUpdate = storageService.readData(
                "shipments",
                `${shipmentId}.json`
              );
              if (shipmentForUpdate) {
                const eventIndex = shipmentForUpdate.events.findIndex(
                  (e) => e.id === event.id
                );
                if (eventIndex >= 0) {
                  shipmentForUpdate.events[
                    eventIndex
                  ].blockchainRecorded = true;
                  shipmentForUpdate.events[eventIndex].blockchainTxHash =
                    result.txHash;
                  storageService.storeData(
                    "shipments",
                    `${shipmentId}.json`,
                    shipmentForUpdate
                  );
                }
              }
            }
          }
        } else if (this.useBridge) {
          // If we're not responsible for this shard but bridge is available,
          // we can try to route it to the correct chain
          logger.info(
            `Shipment ${shipmentId} belongs to another shard, using bridge service`
          );
          // Bridge functionality would be called here
        }
      }

      return event;
    } catch (error) {
      logger.error(`Failed to record event: ${error.message}`);
      throw error;
    }
  }

  /**
   * Process pending events in batch
   * @private
   */
  async _processPendingEvents() {
    // Skip if no pending events
    if (this.pendingEvents.size === 0) {
      return;
    }

    const now = Date.now();
    logger.info(
      `Processing batch of pending events (interval: ${
        now - this.lastBatchTime
      }ms)`
    );
    this.lastBatchTime = now;

    // Create a copy of the pending events map and clear original
    const pendingEventsCopy = new Map(this.pendingEvents);
    this.pendingEvents.clear();

    // Process all pending events
    const promises = [];

    for (const [shipmentId, events] of pendingEventsCopy) {
      for (const event of events) {
        if (event.type === "registerShipment") {
          promises.push(
            this._recordShipmentOnBlockchain(
              event.shipmentId,
              event.metadataHash,
              event.chain
            ).then((result) =>
              this._updateShipmentBlockchainStatus(event.shipmentId, result)
            )
          );
        } else if (event.type === "recordEvent") {
          promises.push(
            this._recordEventOnBlockchain(
              event.shipmentId,
              event.eventType,
              event.metadataHash,
              event.chain
            )
            // We don't update event blockchain status here as we'd need the event ID
          );
        }
      }
    }

    // Wait for all promises to settle
    const results = await Promise.allSettled(promises);

    // Count successes and failures
    const successes = results.filter(
      (r) => r.status === "fulfilled" && r.value?.success
    ).length;
    const failures = results.length - successes;

    logger.info(
      `Batch processing complete: ${successes} successful, ${failures} failed`
    );
  }

  /**
   * Update a shipment's blockchain status
   * @param {string} shipmentId - The shipment ID
   * @param {Object} result - The blockchain operation result
   * @private
   */
  async _updateShipmentBlockchainStatus(shipmentId, result) {
    if (!result.success) {
      return;
    }

    try {
      const shipment = storageService.readData(
        "shipments",
        `${shipmentId}.json`
      );
      if (shipment) {
        shipment.blockchainRecorded = true;
        shipment.blockchainTxHash = result.txHash;
        storageService.storeData("shipments", `${shipmentId}.json`, shipment);
      }
    } catch (error) {
      logger.error(
        `Failed to update shipment blockchain status: ${error.message}`
      );
    }
  }

  /**
   * Record a shipment on the blockchain
   * @param {string} shipmentId - The shipment ID
   * @param {string} metadata - The shipment metadata
   * @param {string} chain - The target blockchain
   * @returns {Promise<{success: boolean, txHash: string, error: string}>}
   * @private
   */
  async _recordShipmentOnBlockchain(shipmentId, metadata, chain = "ethereum") {
    try {
      // Default to Ethereum implementation
      if (chain === "ethereum") {
        return await ethService.registerShipment(shipmentId, metadata);
      } else if (this.useBridge) {
        // Use bridge to handle other chains
        const sourceTxResult = await ethService.registerShipment(
          shipmentId,
          metadata
        );
        if (!sourceTxResult.success) {
          return sourceTxResult;
        }

        // Initiate transfer to target chain
        const transferResult = await bridgeService.initiateTransfer(
          shipmentId,
          "ethereum",
          chain
        );

        return {
          success: transferResult.success,
          txHash: sourceTxResult.txHash,
          transferId: transferResult.transferId,
          error: transferResult.error,
        };
      }

      return {
        success: false,
        txHash: null,
        error: `Chain ${chain} not supported`,
      };
    } catch (error) {
      logger.error(`Failed to record shipment on blockchain: ${error.message}`);

      return {
        success: false,
        txHash: null,
        error: error.message,
      };
    }
  }

  /**
   * Record an event on the blockchain
   * @param {string} shipmentId - The shipment ID
   * @param {string} eventType - The event type
   * @param {string} metadata - The event metadata
   * @param {string} chain - The target blockchain
   * @returns {Promise<{success: boolean, txHash: string, error: string}>}
   * @private
   */
  async _recordEventOnBlockchain(
    shipmentId,
    eventType,
    metadata,
    chain = "ethereum"
  ) {
    try {
      // Default to Ethereum implementation
      if (chain === "ethereum") {
        return await ethService.recordEvent(shipmentId, eventType, metadata);
      } else if (this.useBridge) {
        // Check if shipment exists on target chain
        const verifyResult = await bridgeService._verifyShipmentOnChain(
          shipmentId,
          chain
        );

        if (!verifyResult.exists) {
          // Shipment doesn't exist on target chain, initiate transfer first
          const transferResult = await bridgeService.initiateTransfer(
            shipmentId,
            "ethereum",
            chain
          );

          if (!transferResult.success) {
            return {
              success: false,
              txHash: null,
              error: `Failed to transfer shipment to ${chain}: ${transferResult.error}`,
            };
          }

          // Wait for transfer to complete (this is synchronous for demo, should be async in production)
          // For now, just record on Ethereum
          return await ethService.recordEvent(shipmentId, eventType, metadata);
        } else {
          // Shipment exists on target chain, record event there
          // For now, just record on Ethereum as proof of concept
          return await ethService.recordEvent(shipmentId, eventType, metadata);
        }
      }

      return {
        success: false,
        txHash: null,
        error: `Chain ${chain} not supported`,
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

  /**
   * Get a shipment by ID
   * @param {string} shipmentId - The shipment ID
   * @returns {Object} The shipment
   */
  getShipment(shipmentId) {
    return storageService.readData("shipments", `${shipmentId}.json`);
  }

  /**
   * Verify a shipment on the blockchain
   * @param {string} shipmentId - The shipment ID
   * @returns {Promise<Object>} Verification result
   */
  async verifyShipment(shipmentId) {
    try {
      // First check if sharding is enabled and determine the responsible chain
      let targetChain = "ethereum";
      if (this.useSharding) {
        const shipment = this.getShipment(shipmentId);
        if (shipment && shipment.targetChain) {
          targetChain = shipment.targetChain;
        }
      }

      if (targetChain === "ethereum") {
        return await ethService.verifyShipment(shipmentId);
      } else if (this.useBridge) {
        return await bridgeService._verifyShipmentOnChain(
          shipmentId,
          targetChain
        );
      }

      return {
        exists: false,
        metadata: null,
        registeredBy: null,
        timestamp: null,
        error: `Chain ${targetChain} not supported`,
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

  /**
   * Get shipment events from blockchain
   * @param {string} shipmentId - The shipment ID
   * @returns {Promise<Object>} Events result
   */
  async getBlockchainEvents(shipmentId) {
    try {
      // First check if sharding is enabled and determine the responsible chain
      let targetChain = "ethereum";
      if (this.useSharding) {
        const shipment = this.getShipment(shipmentId);
        if (shipment && shipment.targetChain) {
          targetChain = shipment.targetChain;
        }
      }

      if (targetChain === "ethereum") {
        return await ethService.getEvents(shipmentId);
      } else if (this.useBridge) {
        // Bridge doesn't support getEvents yet, fallback to ethereum
        return await ethService.getEvents(shipmentId);
      }

      return {
        success: false,
        events: [],
        error: `Chain ${targetChain} not supported`,
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

module.exports = new ShipmentService();
