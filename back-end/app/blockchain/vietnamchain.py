from typing import Dict, Any, Optional
import json
import uuid
import hmac
import hashlib
import base64
import time
from datetime import datetime

from app.blockchain.base import BlockchainClient
from app.utils.logger import get_logger
from app.core.exceptions import BlockchainError

logger = get_logger(__name__)

class VietnamChainClient(BlockchainClient):
    """
    Vietnam blockchain client for interacting with the Vietnam national blockchain.
    """
    
    def __init__(
        self, 
        node_url: str,
        api_key: str,
        api_secret: str,
        organization_id: str
    ):
        """
        Initialize the VietnamChain client
        
        Args:
            node_url: URL of the Vietnam blockchain API
            api_key: API key for authentication
            api_secret: API secret for request signing
            organization_id: Registered organization ID on Vietnam blockchain
        """
        super().__init__(node_url, api_key)
        self.api_secret = api_secret
        self.organization_id = organization_id
        logger.info(f"Initialized VietnamChain client for organization {organization_id}")
    
    def _sign_request(self, method: str, endpoint: str, data: Optional[Dict[str, Any]] = None) -> Dict[str, str]:
        """
        Sign a request using HMAC
        
        Args:
            method: HTTP method
            endpoint: API endpoint
            data: Request data
            
        Returns:
            Dictionary with authentication headers
        """
        timestamp = str(int(time.time() * 1000))  # Current time in milliseconds
        
        # Prepare the message to sign
        if data:
            message = f"{method.upper()}:{endpoint}:{json.dumps(data, sort_keys=True)}:{timestamp}"
        else:
            message = f"{method.upper()}:{endpoint}:{timestamp}"
        
        # Create signature
        signature = hmac.new(
            self.api_secret.encode('utf-8'),
            message.encode('utf-8'),
            hashlib.sha256
        ).digest()
        
        signature_b64 = base64.b64encode(signature).decode('utf-8')
        
        # Return headers
        return {
            "X-Auth-ApiKey": self.api_key,
            "X-Auth-Timestamp": timestamp,
            "X-Auth-Signature": signature_b64,
            "X-Organization-ID": self.organization_id
        }
    
    async def register_shipment(
        self, 
        shipment_id: str, 
        tracking_number: str, 
        data_hash: str, 
        metadata: str
    ) -> str:
        """
        Register a shipment on the Vietnam blockchain
        
        Args:
            shipment_id: Unique identifier for the shipment
            tracking_number: Shipment tracking number
            data_hash: Hash of the shipment data
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        endpoint = "api/v1/logistics/shipments"
        
        # Prepare data
        data = {
            "shipmentId": shipment_id,
            "trackingNumber": tracking_number,
            "dataHash": data_hash,
            "metadata": metadata,
            "timestamp": self.current_timestamp()
        }
        
        # Sign request
        headers = self._sign_request("POST", endpoint, data)
        
        # Make request
        response = await self._make_request("POST", endpoint, data, headers)
        
        # Check for errors
        if "error" in response:
            error_msg = f"Failed to register shipment on VietnamChain: {response.get('error')}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        
        # Get transaction hash
        tx_hash = response.get("transactionId")
        if not tx_hash:
            error_msg = "No transaction ID returned from VietnamChain"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        
        logger.info(f"Registered shipment {shipment_id} on VietnamChain, tx_hash: {tx_hash}")
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
        Register an event on the Vietnam blockchain
        
        Args:
            shipment_id: Shipment identifier 
            event_id: Unique identifier for the event
            event_type: Type of event
            data_hash: Hash of the event data
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        endpoint = "api/v1/logistics/events"
        
        # Prepare data
        data = {
            "shipmentId": shipment_id,
            "eventId": event_id,
            "eventType": event_type,
            "dataHash": data_hash,
            "metadata": metadata,
            "timestamp": self.current_timestamp()
        }
        
        # Sign request
        headers = self._sign_request("POST", endpoint, data)
        
        # Make request
        response = await self._make_request("POST", endpoint, data, headers)
        
        # Check for errors
        if "error" in response:
            error_msg = f"Failed to register event on VietnamChain: {response.get('error')}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        
        # Get transaction hash
        tx_hash = response.get("transactionId")
        if not tx_hash:
            error_msg = "No transaction ID returned from VietnamChain"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        
        logger.info(f"Registered event {event_id} for shipment {shipment_id} on VietnamChain, tx_hash: {tx_hash}")
        return tx_hash
    
    async def register_document(
        self, 
        document_hash: str, 
        metadata: str
    ) -> str:
        """
        Register a document hash on the Vietnam blockchain
        
        Args:
            document_hash: Hash of the document
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        endpoint = "api/v1/documents"
        
        # Generate a unique ID for the document
        document_id = str(uuid.uuid4())
        
        # Prepare data
        data = {
            "documentId": document_id,
            "documentHash": document_hash,
            "metadata": metadata,
            "timestamp": self.current_timestamp()
        }
        
        # Sign request
        headers = self._sign_request("POST", endpoint, data)
        
        # Make request
        response = await self._make_request("POST", endpoint, data, headers)
        
        # Check for errors
        if "error" in response:
            error_msg = f"Failed to register document on VietnamChain: {response.get('error')}"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        
        # Get transaction hash
        tx_hash = response.get("transactionId")
        if not tx_hash:
            error_msg = "No transaction ID returned from VietnamChain"
            logger.error(error_msg)
            raise BlockchainError(detail=error_msg)
        
        logger.info(f"Registered document {document_hash} on VietnamChain, tx_hash: {tx_hash}")
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
            Dictionary with transaction status
        """
        endpoint = f"api/v1/transactions/{tx_hash}"
        
        # Sign request
        headers = self._sign_request("GET", endpoint)
        
        # Make request
        response = await self._make_request("GET", endpoint, headers=headers)
        
        # Check for errors
        if "error" in response:
            # If error indicates transaction not found, return appropriate status
            if "not found" in response.get("error", "").lower():
                return {
                    "status": "not_found",
                    "tx_hash": tx_hash,
                    "error": "Transaction not found"
                }
            
            logger.error(f"Failed to get transaction status: {response.get('error')}")
            return {
                "status": "error",
                "tx_hash": tx_hash,
                "error": response.get("error")
            }
        
        # Map VietnamChain status to our standard status
        status_map = {
            "PENDING": "pending",
            "PROCESSING": "pending",
            "CONFIRMED": "confirmed",
            "FAILED": "failed"
        }
        
        # Extract entity type and ID if available
        entity_type = None
        entity_id = None
        
        if "data" in response and "type" in response["data"]:
            entity_type = response["data"]["type"].lower()
            
            if entity_type == "shipment":
                entity_id = response["data"].get("shipmentId")
            elif entity_type == "event":
                entity_id = response["data"].get("eventId")
            elif entity_type == "document":
                entity_id = response["data"].get("documentHash")
        
        # Return standardized response
        return {
            "tx_hash": tx_hash,
            "status": status_map.get(response.get("status", ""), "unknown"),
            "block_hash": response.get("blockHash"),
            "block_number": response.get("blockNumber"),
            "timestamp": response.get("timestamp"),
            "entity_type": entity_type,
            "entity_id": entity_id
        }
    
    async def verify_shipment(
        self, 
        shipment_id: str,
        tracking_number: str
    ) -> Dict[str, Any]:
        """
        Verify a shipment on the Vietnam blockchain
        
        Args:
            shipment_id: Shipment identifier
            tracking_number: Shipment tracking number
            
        Returns:
            Dictionary with verification result
        """
        endpoint = f"api/v1/logistics/shipments/{shipment_id}"
        
        # Sign request
        headers = self._sign_request("GET", endpoint)
        
        # Make request
        response = await self._make_request("GET", endpoint, headers=headers)
        
        # Check for errors
        if "error" in response:
            # If error indicates shipment not found, return appropriate status
            if "not found" in response.get("error", "").lower():
                return {
                    "verified": False,
                    "shipment_id": shipment_id,
                    "tracking_number": tracking_number,
                    "reason": "Shipment not found on blockchain"
                }
            
            logger.error(f"Failed to verify shipment: {response.get('error')}")
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "tracking_number": tracking_number,
                "error": response.get("error")
            }
        
        # Check tracking number match
        stored_tracking = response.get("trackingNumber")
        if stored_tracking != tracking_number:
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "tracking_number": tracking_number,
                "reason": f"Tracking number mismatch (stored: {stored_tracking})"
            }
        
        # Return verification result
        return {
            "verified": True,
            "shipment_id": shipment_id,
            "tracking_number": tracking_number,
            "data_hash": response.get("dataHash"),
            "timestamp": response.get("timestamp"),
            "metadata": response.get("metadata"),
            "tx_hash": response.get("transactionId")
        }
    
    async def verify_event(
        self, 
        shipment_id: str,
        event_id: str
    ) -> Dict[str, Any]:
        """
        Verify an event on the Vietnam blockchain
        
        Args:
            shipment_id: Shipment identifier
            event_id: Event identifier
            
        Returns:
            Dictionary with verification result
        """
        endpoint = f"api/v1/logistics/events/{event_id}"
        
        # Sign request
        headers = self._sign_request("GET", endpoint)
        
        # Make request
        response = await self._make_request("GET", endpoint, headers=headers)
        
        # Check for errors
        if "error" in response:
            # If error indicates event not found, return appropriate status
            if "not found" in response.get("error", "").lower():
                return {
                    "verified": False,
                    "shipment_id": shipment_id,
                    "event_id": event_id,
                    "reason": "Event not found on blockchain"
                }
            
            logger.error(f"Failed to verify event: {response.get('error')}")
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "event_id": event_id,
                "error": response.get("error")
            }
        
        # Check shipment ID match
        stored_shipment_id = response.get("shipmentId")
        if stored_shipment_id != shipment_id:
            return {
                "verified": False,
                "shipment_id": shipment_id,
                "event_id": event_id,
                "reason": f"Shipment ID mismatch (stored: {stored_shipment_id})"
            }
        
        # Return verification result
        return {
            "verified": True,
            "shipment_id": shipment_id,
            "event_id": event_id,
            "event_type": response.get("eventType"),
            "data_hash": response.get("dataHash"),
            "timestamp": response.get("timestamp"),
            "metadata": response.get("metadata"),
            "tx_hash": response.get("transactionId")
        }
    
    async def verify_document(
        self, 
        document_hash: str
    ) -> Dict[str, Any]:
        """
        Verify a document on the Vietnam blockchain
        
        Args:
            document_hash: Document hash
            
        Returns:
            Dictionary with verification result
        """
        endpoint = f"api/v1/documents/hash/{document_hash}"
        
        # Sign request
        headers = self._sign_request("GET", endpoint)
        
        # Make request
        response = await self._make_request("GET", endpoint, headers=headers)
        
        # Check for errors
        if "error" in response:
            # If error indicates document not found, return appropriate status
            if "not found" in response.get("error", "").lower():
                return {
                    "verified": False,
                    "document_hash": document_hash,
                    "reason": "Document not found on blockchain"
                }
            
            logger.error(f"Failed to verify document: {response.get('error')}")
            return {
                "verified": False,
                "document_hash": document_hash,
                "error": response.get("error")
            }
        
        # Return verification result
        return {
            "verified": True,
            "document_hash": document_hash,
            "document_id": response.get("documentId"),
            "timestamp": response.get("timestamp"),
            "metadata": response.get("metadata"),
            "tx_hash": response.get("transactionId")
        }