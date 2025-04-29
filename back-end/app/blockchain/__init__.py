from app.config import settings
from app.utils.logger import get_logger

# Import blockchain clients
from app.blockchain.ethereum import EthereumClient
from app.blockchain.substrate import SubstrateClient
from app.blockchain.vietnamchain import VietnamChainClient

logger = get_logger(__name__)

# Initialize blockchain clients
ethereum_client = None
substrate_client = None
vietnamchain_client = None

# Initialize Ethereum client if enabled
if settings.BLOCKCHAIN_ETHEREUM_ENABLED:
    try:
        ethereum_client = EthereumClient(
            node_url=settings.BLOCKCHAIN_ETHEREUM_NODE_URL,
            private_key=settings.BLOCKCHAIN_ETHEREUM_PRIVATE_KEY,
            shipment_registry_address=settings.BLOCKCHAIN_ETHEREUM_SHIPMENT_REGISTRY_ADDRESS,
            event_log_address=settings.BLOCKCHAIN_ETHEREUM_EVENT_LOG_ADDRESS,
            chain_id=settings.BLOCKCHAIN_ETHEREUM_CHAIN_ID,
            api_key=settings.BLOCKCHAIN_ETHEREUM_API_KEY
        )
        logger.info("Ethereum blockchain client initialized")
    except Exception as e:
        logger.error(f"Failed to initialize Ethereum client: {str(e)}")

# Initialize Substrate client if enabled
if settings.BLOCKCHAIN_SUBSTRATE_ENABLED:
    try:
        substrate_client = SubstrateClient(
            node_url=settings.BLOCKCHAIN_SUBSTRATE_NODE_URL,
            mnemonic=settings.BLOCKCHAIN_SUBSTRATE_MNEMONIC,
            ss58_format=settings.BLOCKCHAIN_SUBSTRATE_SS58_FORMAT,
            api_key=settings.BLOCKCHAIN_SUBSTRATE_API_KEY
        )
        logger.info("Substrate blockchain client initialized")
    except Exception as e:
        logger.error(f"Failed to initialize Substrate client: {str(e)}")

# Initialize VietnamChain client if enabled
if settings.BLOCKCHAIN_VIETNAMCHAIN_ENABLED:
    try:
        vietnamchain_client = VietnamChainClient(
            node_url=settings.BLOCKCHAIN_VIETNAMCHAIN_NODE_URL,
            api_key=settings.BLOCKCHAIN_VIETNAMCHAIN_API_KEY,
            api_secret=settings.BLOCKCHAIN_VIETNAMCHAIN_API_SECRET,
            organization_id=settings.BLOCKCHAIN_VIETNAMCHAIN_ORGANIZATION_ID
        )
        logger.info("VietnamChain blockchain client initialized")
    except Exception as e:
        logger.error(f"Failed to initialize VietnamChain client: {str(e)}")