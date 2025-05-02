const path = require("path");

const env = process.env.NODE_ENV || "development";

const config = {
  NODE: {
    PORT: process.env.PORT || 3000,
    HOST: process.env.HOST || "localhost",
    ENV: env,
    ID: process.env.NODE_ID || 0,
    COUNT: process.env.NODE_COUNT || 1,
  },

  STORAGE: {
    ROOT_DIR: path.join(__dirname, "../../storage"),
  },

  JWT: {
    SECRET:
      process.env.JWT_SECRET ||
      "default-secret-key-for-jwt-replace-in-production",
    EXPIRES_IN: "24h",
  },

  BLOCKCHAIN: {
    ENABLED: process.env.BLOCKCHAIN_ENABLED === "true" || true,
    USE_OPTIMIZED_CONTRACT:
      process.env.USE_OPTIMIZED_CONTRACT === "true" || true,

    CONSENSUS: {
      TYPE: process.env.CONSENSUS_TYPE || "HYBRID",
      VALIDATORS_REQUIRED: parseInt(process.env.VALIDATORS_REQUIRED) || 3,
      VALIDATOR_THRESHOLD: parseFloat(process.env.VALIDATOR_THRESHOLD) || 0.66,
      BLOCK_TIME: parseInt(process.env.BLOCK_TIME) || 5000,
    },

    SHARDING: {
      ENABLED: process.env.SHARDING_ENABLED === "true" || false,
      SHARD_COUNT: parseInt(process.env.SHARD_COUNT) || 16,
      USE_CONTRACT: process.env.SHARDING_USE_CONTRACT === "true" || true,
      VALIDATORS: process.env.SHARD_VALIDATORS
        ? JSON.parse(process.env.SHARD_VALIDATORS)
        : {},
    },

    BRIDGE: {
      ENABLED: process.env.BRIDGE_ENABLED === "true" || false,
      VERIFICATION_THRESHOLD:
        parseFloat(process.env.BRIDGE_VERIFICATION_THRESHOLD) || 0.66,
      RETRY_INTERVAL: parseInt(process.env.BRIDGE_RETRY_INTERVAL) || 300000,
    },

    ETHEREUM: {
      ENABLED: process.env.ETHEREUM_ENABLED === "true" || true,
      NODE_URL: process.env.ETHEREUM_NODE_URL || "http://localhost:8545",
      PRIVATE_KEY: process.env.ETHEREUM_PRIVATE_KEY,
      CONTRACT_ADDRESS: process.env.ETHEREUM_CONTRACT_ADDRESS,
      GAS_LIMIT: process.env.ETHEREUM_GAS_LIMIT || 6000000,
      GAS_PRICE: process.env.ETHEREUM_GAS_PRICE || "20000000000",
      CONSENSUS_TYPE: process.env.ETHEREUM_CONSENSUS_TYPE || "POS",
      VALIDATOR_THRESHOLD:
        parseFloat(process.env.ETHEREUM_VALIDATOR_THRESHOLD) || 0.66,
    },

    POLYGON: {
      ENABLED: process.env.POLYGON_ENABLED === "true" || false,
      NODE_URL: process.env.POLYGON_NODE_URL || "https://polygon-rpc.com",
      PRIVATE_KEY: process.env.POLYGON_PRIVATE_KEY,
      CONTRACT_ADDRESS: process.env.POLYGON_CONTRACT_ADDRESS,
      GAS_LIMIT: process.env.POLYGON_GAS_LIMIT || 6000000,
      GAS_PRICE: process.env.POLYGON_GAS_PRICE || "30000000000",
      CONSENSUS_TYPE: "POS",
    },

    SUBSTRATE: {
      ENABLED: process.env.SUBSTRATE_ENABLED === "true" || false,
      NODE_URL: process.env.SUBSTRATE_NODE_URL || "ws://localhost:9944",
      ACCOUNT_URI: process.env.SUBSTRATE_ACCOUNT_URI,
      CONTRACT_ADDRESS: process.env.SUBSTRATE_CONTRACT_ADDRESS,
      CONSENSUS_TYPE: "NOMINATED_PROOF_OF_STAKE",
    },
  },

  SHIPMENT: {
    FAST_CACHE_ENABLED: process.env.SHIPMENT_FAST_CACHE === "true" || true,
    BLOCKCHAIN_WRITE_INTERVAL:
      parseInt(process.env.BLOCKCHAIN_WRITE_INTERVAL) || 0,
    MAX_BATCH_SIZE: parseInt(process.env.MAX_BATCH_SIZE) || 100,
  },

  LOGGING: {
    LEVEL: process.env.LOG_LEVEL || "info",
    FILE_ERROR: path.join(__dirname, "../../logs/error.log"),
    FILE_COMBINED: path.join(__dirname, "../../logs/combined.log"),
  },
};

const envConfig = {};

switch (env) {
  case "production":
    envConfig.NODE = {
      PORT: process.env.PORT || 80,
    };

    if (!process.env.JWT_SECRET) {
      console.warn("WARNING: JWT_SECRET not set in production environment!");
    }

    if (
      !process.env.ETHEREUM_PRIVATE_KEY &&
      config.BLOCKCHAIN.ETHEREUM.ENABLED
    ) {
      console.warn(
        "WARNING: ETHEREUM_PRIVATE_KEY not set but Ethereum is enabled!"
      );
    }

    break;

  case "test":
    envConfig.BLOCKCHAIN = {
      ENABLED: false,
    };

    break;
}

const mergedConfig = {
  ...config,
  ...envConfig,
  NODE: { ...config.NODE, ...(envConfig.NODE || {}) },
  BLOCKCHAIN: { ...config.BLOCKCHAIN, ...(envConfig.BLOCKCHAIN || {}) },
};

module.exports = mergedConfig;
