# blockchain.py
from sqlalchemy import Column, String, Integer, DateTime, ForeignKey, JSON, Text, Boolean, Float
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID, JSONB
import uuid

from app.db.database import Base
from app.models.base import TimestampMixin, UUIDMixin, BaseCRUD

class BlockchainTransaction(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Model for tracking blockchain transactions"""
    __tablename__ = "blockchain_transactions"

    # Transaction details
    tx_hash = Column(String, index=True, nullable=False)
    network = Column(String, nullable=False)  # ethereum, substrate, vietnam_chain, etc.
    block_number = Column(Integer, nullable=True)
    from_address = Column(String, nullable=False)
    to_address = Column(String, nullable=False)
    value = Column(String, default="0", nullable=False)
    gas_price = Column(String, nullable=True)
    gas_used = Column(Integer, nullable=True)
    
    # Transaction status
    status = Column(String, default="pending", nullable=False)  # pending, confirmed, failed
    error_message = Column(Text, nullable=True)
    confirmations = Column(Integer, default=0, nullable=False)
    
    # Contract interaction
    contract_address = Column(String, nullable=True)
    function_name = Column(String, nullable=True)
    function_args = Column(JSONB, nullable=True)
    
    # Metadata
    metadata = Column(JSON, nullable=True)
    
    # Relationships - what this transaction is related to
    resource_type = Column(String, nullable=True)  # shipment, event, document, etc.
    resource_id = Column(UUID(as_uuid=True), nullable=True)
    
    # User who initiated the transaction (if applicable)
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=True)
    
    def __repr__(self):
        return f"<BlockchainTransaction {self.network}:{self.tx_hash}>"

class BlockchainConfig(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Configuration for blockchain networks"""
    __tablename__ = "blockchain_configs"

    network_name = Column(String, unique=True, nullable=False)
    network_type = Column(String, nullable=False)  # ethereum, substrate, vietnam_chain, etc.
    rpc_url = Column(String, nullable=False)
    chain_id = Column(Integer, nullable=True)
    explorer_url = Column(String, nullable=True)
    
    # Smart contracts
    shipment_contract_address = Column(String, nullable=True)
    event_contract_address = Column(String, nullable=True)
    bridge_contract_address = Column(String, nullable=True)
    
    # Network status
    is_active = Column(Boolean, default=True, nullable=False)
    is_primary = Column(Boolean, default=False, nullable=False)
    last_checked = Column(DateTime(timezone=True), nullable=True)
    
    # Performance metrics
    avg_block_time = Column(Float, nullable=True)  # in seconds
    avg_transaction_fee = Column(Float, nullable=True)  # in native currency
    
    # Network metadata
    metadata = Column(JSON, nullable=True)
    
    def __repr__(self):
        return f"<BlockchainConfig {self.network_name}>"

class Oracle(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Oracle configuration for data fetching"""
    __tablename__ = "oracles"

    name = Column(String, nullable=False)
    oracle_type = Column(String, nullable=False)  # iot, gps, external_api, etc.
    endpoint_url = Column(String, nullable=False)
    auth_type = Column(String, default="api_key", nullable=False)  # api_key, oauth, none, etc.
    
    # Authentication details (encrypted)
    auth_params = Column(JSON, nullable=True)
    
    # Oracle status
    is_active = Column(Boolean, default=True, nullable=False)
    last_check = Column(DateTime(timezone=True), nullable=True)
    last_check_status = Column(String, nullable=True)
    
    # Polling configuration
    polling_interval = Column(Integer, default=300, nullable=False)  # in seconds
    timeout = Column(Integer, default=30, nullable=False)  # in seconds
    
    # Data mapping
    data_mapping = Column(JSON, nullable=True)  # How to map external data to our models
    metadata = Column(JSON, nullable=True)
    
    def __repr__(self):
        return f"<Oracle {self.name}>"

class Bridge(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Bridge configuration for cross-chain operations"""
    __tablename__ = "bridges"

    name = Column(String, nullable=False)
    source_network = Column(String, nullable=False)
    target_network = Column(String, nullable=False)
    bridge_type = Column(String, nullable=False)  # direct, relay, custodial, etc.
    
    # Bridge contract addresses
    source_contract = Column(String, nullable=True)
    target_contract = Column(String, nullable=True)
    
    # Bridge status
    is_active = Column(Boolean, default=True, nullable=False)
    last_check = Column(DateTime(timezone=True), nullable=True)
    last_check_status = Column(String, nullable=True)
    
    # Configuration
    confirmation_blocks = Column(Integer, default=12, nullable=False)
    metadata = Column(JSON, nullable=True)
    
    def __repr__(self):
        return f"<Bridge {self.name}: {self.source_network} -> {self.target_network}>"

class BridgeTransaction(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Transactions processed through bridges"""
    __tablename__ = "bridge_transactions"

    bridge_id = Column(UUID(as_uuid=True), ForeignKey("bridges.id"), nullable=False)
    source_tx_hash = Column(String, nullable=False)
    target_tx_hash = Column(String, nullable=True)
    
    # Status 
    status = Column(String, default="initiated", nullable=False)  # initiated, in_progress, completed, failed
    
    # Data being bridged
    resource_type = Column(String, nullable=False)  # shipment, event, etc.
    resource_id = Column(UUID(as_uuid=True), nullable=False)
    data = Column(JSON, nullable=False)
    
    # Error handling
    error_message = Column(Text, nullable=True)
    retry_count = Column(Integer, default=0, nullable=False)
    
    # Metadata
    metadata = Column(JSON, nullable=True)
    
    def __repr__(self):
        return f"<BridgeTransaction {self.source_tx_hash} -> {self.target_tx_hash}>"