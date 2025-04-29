import asyncio
import json
import os
from typing import Dict, Any, Optional, List, Tuple
from pathlib import Path
import time

import web3
from web3 import Web3
from web3.contract import Contract
from web3.exceptions import TransactionNotFound
from web3.middleware import geth_poa_middleware
from eth_account import Account
from eth_account.signers.local import LocalAccount

from app.blockchain.base import BlockchainClient
from app.blockchain.contracts.shipment_registry import ShipmentRegistry
from app.blockchain.contracts.event_log import EventLog
from app.utils.logger import get_logger
from app.core.exceptions import BlockchainError

logger = get_logger(__name__)

class EthereumClient(BlockchainClient):
    """Client for interacting with Ethereum blockchain"""
    
    def __init__(
        self,
        node_url: str,
        private_key: str,
        shipment_registry_address: Optional[str] = None,
        event_log_address: Optional[str] = None,
        chain_id: int = 1,
        api_key: Optional[str] = None
    ):
        """
        Initialize the Ethereum client
        
        Args:
            node_url: URL of the Ethereum node
            private_key: Private key for signing transactions
            shipment_registry_address: Address of the ShipmentRegistry contract
            event_log_address: Address of the EventLog contract
            chain_id: Ethereum chain ID
            api_key: API key for node provider (if needed)
        """
        super().__init__(node_url, api_key)
        
        # Connect to Ethereum node
        self.web3 = Web3(Web3.HTTPProvider(node_url))
        
        # Add PoA middleware for networks like Goerli
        self.web3.middleware_onion.inject(geth_poa_middleware, layer=0)
        
        # Set up account for signing transactions
        self.account = Account.from_key(private_key)
        self.address = self.account.address
        self.chain_id = chain_id
        
        logger.info(f"Ethereum client initialized with address {self.address}")
        
        # Initialize contract interfaces
        self.shipment_registry = None
        self.event_log = None
        
        if shipment_registry_address:
            self.shipment_registry = ShipmentRegistry(
                self.web3, 
                shipment_registry_address
            )
            logger.info(f"ShipmentRegistry contract connected at {shipment_registry_address}")
        
        if event_log_address:
            self.event_log = EventLog(
                self.web3, 
                event_log_address
            )
            logger.info(f"EventLog contract connected at {event_log_address}")

    async def deploy_contract(
        self, 
        contract_name: str,
        constructor_args: List = None
    ) -> Tuple[str, Contract]:
        """
        Deploy a Solidity contract
        
        Args:
            contract_name: Name of the contract
            constructor_args: Constructor arguments
            
        Returns:
            Tuple of (contract_address, contract_instance)
        """
        # Compile the contract if needed
        abi, bytecode = self.compile_solidity_contract(contract_name)
        
        # Prepare transaction
        contract = self.web3.eth.contract(abi=abi, bytecode=bytecode)
        
        # Build constructor transaction
        if constructor_args:
            tx_constructor = contract.constructor(*constructor_args).build_transaction({
                'from': self.address,
                'nonce': self.web3.eth.get_transaction_count(self.address),
                'gas': 5000000,
                'gasPrice': self.web3.eth.gas_price,
                'chainId': self.chain_id
            })
        else:
            tx_constructor = contract.constructor().build_transaction({
                'from': self.address,
                'nonce': self.web3.eth.get_transaction_count(self.address),
                'gas': 5000000,
                'gasPrice': self.web3.eth.gas_price,
                'chainId': self.chain_id
            })
        
        # Sign transaction
        signed_tx = self.web3.eth.account.sign_transaction(tx_constructor, self.account.key)
        
        # Send transaction
        tx_hash = self.web3.eth.send_raw_transaction(signed_tx.rawTransaction)
        logger.info(f"Contract deployment transaction sent: {tx_hash.hex()}")
        
        # Wait for transaction receipt
        tx_receipt = self.web3.eth.wait_for_transaction_receipt(tx_hash)
        contract_address = tx_receipt.contractAddress
        
        logger.info(f"Contract {contract_name} deployed at {contract_address}")
        
        # Create contract instance
        contract_instance = self.web3.eth.contract(
            address=contract_address,
            abi=abi
        )
        
        # Save deployment info
        deployment_info = {
            "contract_name": contract_name,
            "address": contract_address,
            "tx_hash": tx_hash.hex(),
            "block_number": tx_receipt.blockNumber,
            "deployer": self.address,
            "deployed_at": self.current_timestamp()
        }
        
        # Save deployment info to file
        deployment_dir = Path(__file__).parent / "contracts" / "deployments"
        deployment_dir.mkdir(exist_ok=True)
        
        deployment_file = deployment_dir / f"{contract_name}_{self.chain_id}.json"
        with open(deployment_file, 'w') as f:
            json.dump(deployment_info, f, indent=2)
        
        return contract_address, contract_instance
    
    def _sign_and_send_transaction(self, transaction: Dict[str, Any]) -> str:
        """
        Sign and send a transaction
        
        Args:
            transaction: Transaction dictionary
            
        Returns:
            Transaction hash
        """
        # Sign transaction
        signed_tx = self.web3.eth.account.sign_transaction(transaction, self.account.key)
        
        # Send transaction
        tx_hash = self.web3.eth.send_raw_transaction(signed_tx.rawTransaction)
        
        return tx_hash.hex()
    
    async def register_shipment(
        self,
        shipment_id: str,
        tracking_number: str,
        data_hash: str,
        metadata: str
    ) -> str:
        """
        Register a shipment on the blockchain
        
        Args:
            shipment_id: Unique identifier for the shipment
            tracking_number: Tracking number
            data_hash: Hash of the shipment data
            metadata: Additional metadata
            
        Returns:
            Transaction hash
        """
        if not self.shipment_registry:
            raise ValueError("ShipmentRegistry contract not initialized")
        
        # Build transaction
        tx = self.shipment_registry.build_register_shipment_transaction(
            shipment_id=shipment_id,
            tracking_number=tracking_number,
            data_hash=data_hash,
            metadata=metadata,
            sender=self.address
        )
        
        # Sign and send transaction
        tx_hash = self._sign_and_send_transaction(tx)
        
        logger.info(f"Shipment {shipment_id} registered with tx_hash: {tx_hash}")
        
        return tx_hash
    
    async def register_event(
        self,
        shipment_id: str,
        event_id: str,
        event_type: str,
        data_hash: str,
        metadata: str
    ) -> str:
        """
        Register an event on the blockchain
        
        Args:
            shipment_id: Associated shipment ID
            event_id: Unique identifier for the event
            event_type: Type of event
            data_hash: Hash of the event data
            metadata: Additional metadata
            
        Returns:
            Transaction hash
        """
        if not self.event_log:
            raise ValueError("EventLog contract not initialized")
        
        # Build transaction
        tx = self.event_log.build_log_event_transaction(
            shipment_id=shipment_id,
            event_id=event_id,
            event_type=event_type,
            data_hash=data_hash,
            metadata=metadata,
            sender=self.address
        )
        
        # Sign and send transaction
        tx_hash = self._sign_and_send_transaction(tx)
        
        logger.info(f"Event {event_id} for shipment {shipment_id} registered with tx_hash: {tx_hash}")
        
        return tx_hash
    
    async def register_document(
        self,
        document_hash: str,
        metadata: str
    ) -> str:
        """
        Register a document hash on the blockchain
        
        Args:
            document_hash: Hash of the document
            metadata: Additional metadata
            
        Returns:
            Transaction hash
        """
        if not self.event_log:
            raise ValueError("EventLog contract not initialized")
        
        # Generate a document ID
        document_id = f"doc_{int(time.time())}_{document_hash[:8]}"
        
        # Build transaction
        tx = self.event_log.build_log_document_transaction(
            document_id=document_id,
            document_hash=document_hash,
            metadata=metadata,
            sender=self.address
        )
        
        # Sign and send transaction
        tx_hash = self._sign_and_send_transaction(tx)
        
        logger.info(f"Document {document_hash} registered with tx_hash: {tx_hash}")
        
        return tx_hash
    
    async def get_transaction_status(
        self,
        tx_hash: str
    ) -> Dict[str, Any]:
        """
        Get the status of a transaction
        
        Args:
            tx_hash: Transaction hash
            
        Returns:
            Transaction status
        """
        try:
            # Get transaction receipt
            receipt = self.web3.eth.get_transaction_receipt(tx_hash)
            
            if receipt is None:
                # Transaction is pending
                return {
                    "status": "pending",
                    "tx_hash": tx_hash,
                    "confirmations": 0,
                    "network": self._get_network_name()
                }
            
            # Calculate confirmations
            current_block = self.web3.eth.block_number
            confirmations = current_block - receipt.blockNumber
            
            # Get status
            status = "confirmed" if receipt.status == 1 else "failed"
            
            # Get logs
            logs = receipt.logs
            
            # Try to determine entity type and ID from logs
            entity_type = None
            entity_id = None
            
            if self.shipment_registry:
                # Check if any of the logs are from ShipmentRegistry contract
                shipment_logs = self.shipment_registry.decode_logs(logs)
                if shipment_logs:
                    log = shipment_logs[0]
                    entity_type = "shipment"
                    entity_id = log.args.shipmentId
            
            if not entity_id and self.event_log:
                # Check if any of the logs are from EventLog contract
                event_logs = self.event_log.decode_logs(logs)
                for log in event_logs:
                    if hasattr(log.args, 'eventId'):
                        entity_type = "event"
                        entity_id = log.args.eventId
                        break
                    elif hasattr(log.args, 'documentHash'):
                        entity_type = "document"
                        entity_id = log.args.documentHash
                        break
            
            return {
                "status": status,
                "tx_hash": tx_hash,
                "block_number": receipt.blockNumber,
                "confirmations": confirmations,
                "gas_used": receipt.gasUsed,
                "network": self._get_network_name(),
                "entity_type": entity_type,
                "entity_id": entity_id
            }
            
        except TransactionNotFound:
            # Transaction not found
            return {
                "status": "not_found",
                "tx_hash": tx_hash,
                "network": self._get_network_name()
            }
        except Exception as e:
            logger.error(f"Error getting transaction status: {str(e)}")
            return {
                "status": "error",
                "tx_hash": tx_hash,
                "error": str(e),
                "network": self._get_network_name()
            }
    
    async def verify_shipment(
        self,
        shipment_id: str,
        tracking_number: str
    ) -> Dict[str, Any]:
        """
        Verify shipment data on the blockchain
        
        Args:
            shipment_id: Shipment ID
            tracking_number: Tracking number
            
        Returns:
            Verification result
        """
        if not self.shipment_registry:
            raise ValueError("ShipmentRegistry contract not initialized")
        
        try:
            # Get shipment from contract
            result = self.shipment_registry.get_shipment(shipment_id)
            
            # Parse result
            stored_tracking_number, stored_data_hash, timestamp, metadata = result
            
            # Verify tracking number
            tracking_match = stored_tracking_number == tracking_number
            
            # Get additional data
            metadata_json = json.loads(metadata) if metadata else {}
            
            return {
                "verified": tracking_match,
                "shipment_id": shipment_id,
                "tracking_number": tracking_number,
                "stored_tracking_number": stored_tracking_number,
                "data_hash": stored_data_hash,
                "timestamp": timestamp,
                "metadata": metadata_json,
                "network": self._get_network_name()
            }
            
        except Exception as e:
            logger.error(f"Error verifying shipment: {str(e)}")
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "error": str(e),
                "network": self._get_network_name()
            }
    
    async def verify_event(
        self,
        shipment_id: str,
        event_id: str
    ) -> Dict[str, Any]:
        """
        Verify event data on the blockchain
        
        Args:
            shipment_id: Shipment ID
            event_id: Event ID
            
        Returns:
            Verification result
        """
        if not self.event_log:
            raise ValueError("EventLog contract not initialized")
        
        try:
            # Get event from contract
            result = self.event_log.get_event(event_id)
            
            # Parse result
            stored_shipment_id, event_type, data_hash, timestamp, metadata = result
            
            # Verify shipment ID
            shipment_match = stored_shipment_id == shipment_id
            
            # Get additional data
            metadata_json = json.loads(metadata) if metadata else {}
            
            return {
                "verified": shipment_match,
                "shipment_id": shipment_id,
                "event_id": event_id,
                "stored_shipment_id": stored_shipment_id,
                "event_type": event_type,
                "data_hash": data_hash,
                "timestamp": timestamp,
                "metadata": metadata_json,
                "network": self._get_network_name()
            }
            
        except Exception as e:
            logger.error(f"Error verifying event: {str(e)}")
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "event_id": event_id,
                "error": str(e),
                "network": self._get_network_name()
            }
    
    async def verify_document(
        self,
        document_hash: str
    ) -> Dict[str, Any]:
        """
        Verify document hash on the blockchain
        
        Args:
            document_hash: Document hash
            
        Returns:
            Verification result
        """
        if not self.event_log:
            raise ValueError("EventLog contract not initialized")
        
        try:
            # Get document from contract
            result = self.event_log.get_document(document_hash)
            
            # Parse result
            timestamp, metadata = result
            
            # Get additional data
            metadata_json = json.loads(metadata) if metadata else {}
            
            return {
                "verified": True,
                "document_hash": document_hash,
                "timestamp": timestamp,
                "metadata": metadata_json,
                "network": self._get_network_name()
            }
            
        except Exception as e:
            logger.error(f"Error verifying document: {str(e)}")
            return {
                "verified": False,
                "document_hash": document_hash,
                "error": str(e),
                "network": self._get_network_name()
            }
    
    def _get_network_name(self) -> str:
        """Get the name of the connected Ethereum network"""
        if self.chain_id == 1:
            return "ethereum"
        elif self.chain_id == 5:
            return "goerli"
        elif self.chain_id == 11155111:
            return "sepolia"
        elif self.chain_id == 137:
            return "polygon"
        elif self.chain_id == 80001:
            return "polygon_mumbai"
        else:
            return f"chain_{self.chain_id}"
            
    async def get_contract_deployments(self) -> Dict[str, Any]:
        """
        Get information about deployed contracts
        
        Returns:
            Dictionary with contract deployment information
        """
        deployment_dir = Path(__file__).parent / "contracts" / "deployments"
        
        if not deployment_dir.exists():
            return {"contracts": []}
        
        deployments = []
        for file in deployment_dir.glob(f"*_{self.chain_id}.json"):
            try:
                with open(file, 'r') as f:
                    deployment = json.load(f)
                    deployments.append(deployment)
            except Exception as e:
                logger.error(f"Error loading deployment file {file}: {str(e)}")
        
        return {"contracts": deployments}