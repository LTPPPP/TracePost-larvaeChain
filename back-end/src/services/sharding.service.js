const config = require("../config");
const logger = require("../utils/logger");
const ethService = require("../blockchain/ethereum.service");

class ShardingService {
  constructor() {
    this.initialized = false;
    this.shardCount = config.BLOCKCHAIN.SHARDING.SHARD_COUNT || 16;
    this.shardValidators = new Map();
    this.shardingEnabled = config.BLOCKCHAIN.SHARDING.ENABLED || false;

    if (this.shardingEnabled) {
      this.initialize();
    }
  }

  initialize() {
    try {
      logger.info(
        `Initializing sharding service with ${this.shardCount} shards`
      );

      if (config.BLOCKCHAIN.SHARDING.VALIDATORS) {
        for (const [shardId, validator] of Object.entries(
          config.BLOCKCHAIN.SHARDING.VALIDATORS
        )) {
          this.shardValidators.set(parseInt(shardId), validator);
        }
        logger.info(
          `Initialized ${this.shardValidators.size} shard validators`
        );
      }

      this.initialized = true;
      logger.info("Sharding service initialized successfully");
    } catch (error) {
      logger.error(`Sharding service initialization failed: ${error.message}`);
      this.initialized = false;
    }
  }

  getShipmentShard(shipmentId) {
    if (!this.shardingEnabled) return 0;

    const hash = Buffer.from(shipmentId).reduce((a, b) => a + b, 0);
    return hash % this.shardCount;
  }

  isResponsibleForShard(shardId) {
    if (!this.shardingEnabled) return true;

    if (this.shardValidators.has(shardId)) {
      return this.shardValidators.get(shardId) === ethService.getNodeAddress();
    }

    const nodeId = config.NODE.ID || 0;
    return shardId % config.NODE.COUNT === nodeId;
  }

  async registerShardValidator(shardId, validatorAddress) {
    try {
      if (!this.shardingEnabled) {
        logger.warn("Sharding is not enabled, ignoring validator registration");
        return false;
      }

      if (shardId >= this.shardCount) {
        logger.error(
          `Invalid shard ID: ${shardId}, max is ${this.shardCount - 1}`
        );
        return false;
      }

      this.shardValidators.set(shardId, validatorAddress);

      if (
        ethService.isInitialized() &&
        config.BLOCKCHAIN.SHARDING.USE_CONTRACT
      ) {
        await ethService.contract.assignShardValidator(
          shardId,
          validatorAddress
        );
        logger.info(
          `Registered validator ${validatorAddress} for shard ${shardId} in contract`
        );
      }

      logger.info(
        `Registered validator ${validatorAddress} for shard ${shardId}`
      );
      return true;
    } catch (error) {
      logger.error(`Failed to register shard validator: ${error.message}`);
      return false;
    }
  }

  getShardValidators() {
    return this.shardValidators;
  }

  isInitialized() {
    return this.initialized;
  }
}

module.exports = new ShardingService();
