from typing import Dict, Any, Optional, List
import json
import uuid
import asyncio
from substrateinterface import SubstrateInterface, Keypair
from substrateinterface.exceptions import SubstrateRequestException

from app.blockchain.base import BlockchainClient
from app.utils.logger import get_logger
from app.core.exceptions import BlockchainError

logger = get_logger(__name__)

class SubstrateClient(BlockchainClient):
    """
    Substrate blockchain client for interacting with Substrate-based blockchains like Polkadot, Kusama, etc.
    """
    
    def __init__(
        self, 
        node_url: str, 
        mnemonic: str,
        ss58_format: int = 42,
        type_registry: Optional[Dict] = None,
        metadata_cache: Optional[Dict] = None,
        api_key: Optional[str] = None
    ):
        """
        Initialize the Substrate client
        
        Args:
            node_url: URL of the Substrate node
            mnemonic: Mnemonic seed phrase for the account
            ss58_format: SS58 address format
            type_registry: Custom type registry
            metadata_cache: Metadata cache
            api_key: Optional API key for node access
        """
        super().__init__(node_url, api_key)
        
        # Initialize substrate interface
        try:
            self.substrate = SubstrateInterface(
                url=node_url,
                ss58_format=ss58_format,
                type_registry=type_registry,
                type_registry_preset="default",
                cache_region="substrate_cache",
                runtime_config=metadata_cache
            )
        except Exception as e:
            error_msg = f"Failed to connect to Substrate node at {node_url}: {str(e)}"
            logger.error(error_msg)
            raise ConnectionError(error_msg)
        
        # Initialize keypair
        self.keypair = Keypair.create_from_mnemonic(mnemonic, ss58_format=ss58_format)
        self.address = self.keypair.ss58_address
        
        logger.info(f"Connected to Substrate node at {node_url}")
        logger.info(f"Using account: {self.address}")
    
    async def register_shipment(
        self, 
        shipment_id: str, 
        tracking_number: str, 
        data_hash: str, 
        metadata: str
    ) -> str:
        """
        Register a shipment on the Substrate blockchain
        
        Args:
            shipment_id: Unique identifier for the shipment
            tracking_number: Shipment tracking number
            data_hash: Hash of the shipment data
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        try:
            # Create call
            call = self.substrate.compose_call(
                call_module='LogisticsTraceability',
                call_function='register_shipment',
                call_params={
                    'shipment_id': shipment_id,
                    'tracking_number': tracking_number,
                    'data_hash': data_hash,
                    'metadata': metadata
                }
            )
            
            # Create extrinsic
            extrinsic = self.substrate.create_signed_extrinsic(
                call=call, 
                keypair=self.keypair
            )
            
            # Submit extrinsic
            response = self.substrate.submit_extrinsic(
                extrinsic=extrinsic,
                wait_for_inclusion=True
            )
            
            tx_hash = response.extrinsic_hash
            
            logger.info(f"Registered shipment {shipment_id} on Substrate, tx_hash: {tx_hash}")
            return tx_hash
            
        except SubstrateRequestException as e:
            error_msg = f"Substrate request error: {str(e)}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        except Exception as e:
            error_msg = f"Failed to register shipment on Substrate: {str(e)}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
    
    async def register_event(
        self, 
        shipment_id: str, 
        event_id: str, 
        event_type: str, 
        data_hash: str, 
        metadata: str
    ) -> str:
        """
        Register an event on the Substrate blockchain
        
        Args:
            shipment_id: Shipment identifier 
            event_id: Unique identifier for the event
            event_type: Type of event
            data_hash: Hash of the event data
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        try:
            # Create call
            call = self.substrate.compose_call(
                call_module='LogisticsTraceability',
                call_function='register_event',
                call_params={
                    'shipment_id': shipment_id,
                    'event_id': event_id,
                    'event_type': event_type,
                    'data_hash': data_hash,
                    'metadata': metadata
                }
            )
            
            # Create extrinsic
            extrinsic = self.substrate.create_signed_extrinsic(
                call=call, 
                keypair=self.keypair
            )
            
            # Submit extrinsic
            response = self.substrate.submit_extrinsic(
                extrinsic=extrinsic,
                wait_for_inclusion=True
            )
            
            tx_hash = response.extrinsic_hash
            
            logger.info(f"Registered event {event_id} for shipment {shipment_id} on Substrate, tx_hash: {tx_hash}")
            return tx_hash
            
        except SubstrateRequestException as e:
            error_msg = f"Substrate request error: {str(e)}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        except Exception as e:
            error_msg = f"Failed to register event on Substrate: {str(e)}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
    
    async def register_document(
        self, 
        document_hash: str, 
        metadata: str
    ) -> str:
        """
        Register a document hash on the Substrate blockchain
        
        Args:
            document_hash: Hash of the document
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        try:
            # Generate a unique ID for the document
            document_id = str(uuid.uuid4())
            
            # Create call
            call = self.substrate.compose_call(
                call_module='LogisticsTraceability',
                call_function='register_document',
                call_params={
                    'document_id': document_id,
                    'document_hash': document_hash,
                    'metadata': metadata
                }
            )
            
            # Create extrinsic
            extrinsic = self.substrate.create_signed_extrinsic(
                call=call, 
                keypair=self.keypair
            )
            
            # Submit extrinsic
            response = self.substrate.submit_extrinsic(
                extrinsic=extrinsic,
                wait_for_inclusion=True
            )
            
            tx_hash = response.extrinsic_hash
            
            logger.info(f"Registered document {document_hash} on Substrate, tx_hash: {tx_hash}")
            return tx_hash
            
        except SubstrateRequestException as e:
            error_msg = f"Substrate request error: {str(e)}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        except Exception as e:
            error_msg = f"Failed to register document on Substrate: {str(e)}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
    
    async def get_transaction_status(
        self, 
        tx_hash: str
    ) -> Dict[str, Any]:
        """
        Get the status of a transaction
        
        Args:
            tx_hash: Transaction hash
            
        Returns:
            Dictionary with transaction status
        """
        try:
            # Check if the transaction exists
            extrinsic = self.substrate.get_extrinsic(tx_hash)
            
            if not extrinsic:
                return {
                    "status": "not_found",
                    "tx_hash": tx_hash,
                    "error": "Transaction not found"
                }
            
            # Get block information
            block_hash = extrinsic.get('block_hash')
            block_number = extrinsic.get('block_number')
            
            # Get status
            if extrinsic.get('success', False):
                status = "confirmed"
            else:
                status = "failed"
            
            # Get current block
            current_block = self.substrate.get_block_number(self.substrate.get_chain_head())
            confirmations = current_block - block_number if block_number else 0
            
            # Get block timestamp if available
            timestamp = None
            try:
                block = self.substrate.get_block(block_hash)
                for extrinsic in block['extrinsics']:
                    if extrinsic.module_id == 'Timestamp' and extrinsic.call_function == 'set':
                        timestamp = extrinsic.params[0].value
                        break
            except Exception as e:
                logger.warning(f"Failed to get block timestamp: {str(e)}")
            
            # Try to extract entity type and ID from events
            entity_type = None
            entity_id = None
            
            try:
                for event in extrinsic.get('events', []):
                    if event.module_id == 'LogisticsTraceability':
                        if event.event_id == 'ShipmentRegistered':
                            entity_type = "shipment"
                            entity_id = event.params[0].value
                            break
                        elif event.event_id == 'EventRegistered':
                            entity_type = "event"
                            entity_id = event.params[1].value
                            break
                        elif event.event_id == 'DocumentRegistered':
                            entity_type = "document"
                            entity_id = event.params[1].value
                            break
            except Exception as e:
                logger.warning(f"Failed to parse extrinsic events: {str(e)}")
            
            return {
                "tx_hash": tx_hash,
                "status": status,
                "block_hash": block_hash,
                "block_number": block_number,
                "timestamp": timestamp,
                "confirmations": confirmations,
                "entity_type": entity_type,
                "entity_id": entity_id
            }
            
        except Exception as e:
            logger.error(f"Failed to get transaction status: {str(e)}")
            return {
                "status": "error",
                "tx_hash": tx_hash,
                "error": str(e)
            }
    
    async def verify_shipment(
        self, 
        shipment_id: str,
        tracking_number: str
    ) -> Dict[str, Any]:
        """
        Verify a shipment on the Substrate blockchain
        
        Args:
            shipment_id: Shipment identifier
            tracking_number: Shipment tracking number
            
        Returns:
            Dictionary with verification result
        """
        try:
            # Query chain state
            result = self.substrate.query(
                module='LogisticsTraceability',
                storage_function='Shipments',
                params=[shipment_id]
            )
            
            # Check if shipment exists
            if not result or result.value is None:
                return {
                    "verified": False,
                    "shipment_id": shipment_id,
                    "tracking_number": tracking_number,
                    "reason": "Shipment not found on blockchain"
                }
            
            # Extract shipment data
            shipment_data = result.value
            
            # Check tracking number
            stored_tracking = shipment_data.get('tracking_number', '')
            if stored_tracking != tracking_number:
                return {
                    "verified": False,
                    "shipment_id": shipment_id,
                    "tracking_number": tracking_number,
                    "reason": f"Tracking number mismatch (stored: {stored_tracking})"
                }
            
            # Return success
            return {
                "verified": True,
                "shipment_id": shipment_id,
                "tracking_number": tracking_number,
                "data_hash": shipment_data.get('data_hash', ''),
                "timestamp": shipment_data.get('timestamp', 0),
                "metadata": shipment_data.get('metadata', ''),
                "registrar": shipment_data.get('registrar', '')
            }
            
        except Exception as e:
            error_msg = f"Failed to verify shipment: {str(e)}"
            logger.error(error_msg)
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "tracking_number": tracking_number,
                "error": str(e)
            }
    
    async def verify_event(
        self, 
        shipment_id: str,
        event_id: str
    ) -> Dict[str, Any]:
        """
        Verify an event on the Substrate blockchain
        
        Args:
            shipment_id: Shipment identifier
            event_id: Event identifier
            
        Returns:
            Dictionary with verification result
        """
        try:
            # Query chain state
            result = self.substrate.query(
                module='LogisticsTraceability',
                storage_function='Events',
                params=[event_id]
            )
            
            # Check if event exists
            if not result or result.value is None:
                return {
                    "verified": False,
                    "shipment_id": shipment_id,
                    "event_id": event_id,
                    "reason": "Event not found on blockchain"
                }
            
            # Extract event data
            event_data = result.value
            
            # Check shipment ID
            stored_shipment_id = event_data.get('shipment_id', '')
            if stored_shipment_id != shipment_id:
                return {
                    "verified": False,
                    "shipment_id": shipment_id,
                    "event_id": event_id,
                    "reason": f"Shipment ID mismatch (stored: {stored_shipment_id})"
                }
            
            # Return success
            return {
                "verified": True,
                "shipment_id": shipment_id,
                "event_id": event_id,
                "event_type": event_data.get('event_type', ''),
                "data_hash": event_data.get('data_hash', ''),
                "timestamp": event_data.get('timestamp', 0),
                "metadata": event_data.get('metadata', ''),
                "registrar": event_data.get('registrar', '')
            }
            
        except Exception as e:
            error_msg = f"Failed to verify event: {str(e)}"
            logger.error(error_msg)
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "event_id": event_id,
                "error": str(e)
            }
    
    async def verify_document(
        self, 
        document_hash: str
    ) -> Dict[str, Any]:
        """
        Verify a document on the Substrate blockchain
        
        Args:
            document_hash: Document hash
            
        Returns:
            Dictionary with verification result
        """
        try:
            # Query chain state
            result = self.substrate.query_map(
                module='LogisticsTraceability',
                storage_function='Documents'
            )
            
            # Search for document by hash
            found_document = None
            for document_id, document_data in result:
                if document_data.value.get('document_hash') == document_hash:
                    found_document = {
                        'document_id': document_id.value,
                        **document_data.value
                    }
                    break
            
            # Check if document exists
            if not found_document:
                return {
                    "verified": False,
                    "document_hash": document_hash,
                    "reason": "Document not found on blockchain"
                }
            
            # Return success
            return {
                "verified": True,
                "document_hash": document_hash,
                "document_id": found_document.get('document_id', ''),
                "timestamp": found_document.get('timestamp', 0),
                "metadata": found_document.get('metadata', ''),
                "registrar": found_document.get('registrar', '')
            }
            
        except Exception as e:
            error_msg = f"Failed to verify document: {str(e)}"
            logger.error(error_msg)
            return {
                "verified": False,
                "document_hash": document_hash,
                "error": str(e)
            }