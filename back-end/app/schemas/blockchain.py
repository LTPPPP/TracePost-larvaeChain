from typing import Dict, List, Any, Optional
from enum import Enum
from datetime import datetime
from pydantic import BaseModel, Field, validator


class BlockchainNetwork(str, Enum):
    """Supported blockchain networks"""
    ETHEREUM = "ethereum"
    SUBSTRATE = "substrate"
    VIETNAMCHAIN = "vietnamchain"


class BlockchainStatus(BaseModel):
    """Blockchain node status response"""
    enabled: bool
    node_url: Optional[str] = None
    network: Optional[str] = None
    chain_id: Optional[int] = None
    account: Optional[str] = None


class BlockchainSystemStatus(BaseModel):
    """Overall blockchain system status"""
    ethereum: BlockchainStatus
    substrate: BlockchainStatus
    vietnamchain: BlockchainStatus
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class ContractDeploymentRequest(BaseModel):
    """Request to deploy a smart contract"""
    contract_name: str = Field(..., description="Name of the contract to deploy")
    network: BlockchainNetwork = Field(default=BlockchainNetwork.ETHEREUM, description="Blockchain network")
    compiler_version: str = Field(default="0.8.17", description="Solidity compiler version")
    constructor_args: Optional[List[Any]] = Field(default=None, description="Arguments for the contract constructor")
    deploy_in_background: bool = Field(default=False, description="Deploy the contract in a background task")
    
    @validator('contract_name')
    def validate_contract_name(cls, v):
        allowed_contracts = {"ShipmentRegistry", "EventLog", "AccessControl", "SensorDataLog"}
        if v not in allowed_contracts:
            raise ValueError(f"Contract {v} not supported. Allowed contracts: {allowed_contracts}")
        return v


class ContractDeploymentResponse(BaseModel):
    """Result of contract deployment"""
    status: str
    contract_name: str
    contract_address: Optional[str] = None
    network: str
    message: str
    transaction_hash: Optional[str] = None


class VerificationRequest(BaseModel):
    """Request to verify data on blockchain"""
    network: BlockchainNetwork = Field(default=BlockchainNetwork.ETHEREUM, description="Blockchain network")
    include_metadata: bool = Field(default=True, description="Include metadata in response")


class TransactionStatusResponse(BaseModel):
    """Transaction status response"""
    status: str
    tx_hash: str
    block_number: Optional[int] = None
    confirmations: Optional[int] = None
    gas_used: Optional[int] = None
    network: str
    entity_type: Optional[str] = None
    entity_id: Optional[str] = None
    error: Optional[str] = None