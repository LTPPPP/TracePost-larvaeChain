from abc import ABC, abstractmethod
from typing import Dict, Any, Optional, List, Tuple
import os
import json
from pathlib import Path
import solcx  # For compiling Solidity contracts
import aiohttp
import logging
import asyncio
from datetime import datetime

from app.utils.logger import get_logger

logger = get_logger(__name__)

class BlockchainClient(ABC):
    """
    Abstract base class for blockchain clients.
    
    All blockchain implementations should inherit from this class
    and implement the required methods.
    """
    
    def __init__(self, node_url: str, api_key: Optional[str] = None):
        """
        Initialize the blockchain client
        
        Args:
            node_url: URL of the blockchain node
            api_key: Optional API key for authentication
        """
        self.node_url = node_url
        self.api_key = api_key
        self.logger = logging.getLogger(self.__class__.__name__)
        logger.info(f"Initializing blockchain client for {node_url}")
    
    @abstractmethod
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
            tracking_number: Shipment tracking number
            data_hash: Hash of the shipment data
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        pass
    
    @abstractmethod
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
            shipment_id: Shipment identifier 
            event_id: Unique identifier for the event
            event_type: Type of event
            data_hash: Hash of the event data
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        pass
    
    @abstractmethod
    async def register_document(
        self, 
        document_hash: str, 
        metadata: str
    ) -> str:
        """
        Register a document hash on the blockchain
        
        Args:
            document_hash: Hash of the document
            metadata: JSON string with additional metadata
            
        Returns:
            Transaction hash
        """
        pass
    
    @abstractmethod
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
        pass
    
    @abstractmethod
    async def verify_shipment(
        self, 
        shipment_id: str,
        tracking_number: str
    ) -> Dict[str, Any]:
        """
        Verify a shipment on the blockchain
        
        Args:
            shipment_id: Shipment identifier
            tracking_number: Shipment tracking number
            
        Returns:
            Dictionary with verification result
        """
        pass
    
    @abstractmethod
    async def verify_event(
        self, 
        shipment_id: str,
        event_id: str
    ) -> Dict[str, Any]:
        """
        Verify an event on the blockchain
        
        Args:
            shipment_id: Shipment identifier
            event_id: Event identifier
            
        Returns:
            Dictionary with verification result
        """
        pass
    
    @abstractmethod
    async def verify_document(
        self, 
        document_hash: str
    ) -> Dict[str, Any]:
        """
        Verify a document on the blockchain
        
        Args:
            document_hash: Document hash
            
        Returns:
            Dictionary with verification result
        """
        pass
    
    async def _make_request(
        self, 
        method: str, 
        endpoint: str, 
        data: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, str]] = None
    ) -> Dict[str, Any]:
        """
        Make an HTTP request to the blockchain node
        
        Args:
            method: HTTP method (GET, POST, etc.)
            endpoint: API endpoint
            data: Request data
            headers: Request headers
            
        Returns:
            Response data
        """
        url = f"{self.node_url}/{endpoint.lstrip('/')}"
        
        if not headers:
            headers = {}
        
        # Add API key if provided
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        
        headers["Content-Type"] = "application/json"
        
        try:
            async with aiohttp.ClientSession() as session:
                if method.upper() == "GET":
                    async with session.get(url, headers=headers) as response:
                        if response.status != 200:
                            error_text = await response.text()
                            self.logger.error(f"Request failed: {response.status} - {error_text}")
                            return {"error": f"Request failed: {response.status}", "details": error_text}
                        
                        return await response.json()
                        
                elif method.upper() == "POST":
                    async with session.post(url, headers=headers, json=data) as response:
                        if response.status not in (200, 201, 202):
                            error_text = await response.text()
                            self.logger.error(f"Request failed: {response.status} - {error_text}")
                            return {"error": f"Request failed: {response.status}", "details": error_text}
                        
                        return await response.json()
                        
                else:
                    self.logger.error(f"Unsupported HTTP method: {method}")
                    return {"error": f"Unsupported HTTP method: {method}"}
                    
        except aiohttp.ClientError as e:
            self.logger.error(f"Request error: {str(e)}")
            return {"error": f"Request error: {str(e)}"}
        except json.JSONDecodeError:
            self.logger.error(f"Invalid JSON response")
            return {"error": "Invalid JSON response"}
        except Exception as e:
            self.logger.error(f"Unexpected error: {str(e)}")
            return {"error": f"Unexpected error: {str(e)}"}
    
    @staticmethod
    def current_timestamp() -> str:
        """Get current UTC timestamp in ISO format"""
        return datetime.utcnow().isoformat()
    
    def compile_solidity_contract(
        self, 
        contract_name: str,
        solidity_version: str = "0.8.17"
    ) -> Tuple[Dict, str]:
        """
        Compile a Solidity contract
        
        Args:
            contract_name: Name of the contract file (without .sol)
            solidity_version: Solidity compiler version
            
        Returns:
            Tuple of (contract_interface, contract_bytecode)
        """
        try:
            # Install solc version if not installed
            if not solcx.get_installed_solc_versions() or solidity_version not in solcx.get_installed_solc_versions():
                solcx.install_solc(solidity_version)
            
            # Set solc version
            solcx.set_solc_version(solidity_version)
            
            # Get contract path
            contract_file = f"{contract_name}.sol"
            contract_path = Path(__file__).parent / "contracts" / "solidity" / contract_file
            
            # Compile the contract
            compiled_sol = solcx.compile_files(
                [contract_path],
                output_values=["abi", "bin"],
                solc_version=solidity_version
            )
            
            # Get contract interface and bytecode
            contract_id = f"{contract_path}:{contract_name}"
            contract_interface = compiled_sol[contract_id]['abi']
            contract_bytecode = compiled_sol[contract_id]['bin']
            
            # Save ABI to a JSON file
            abi_path = Path(__file__).parent / "contracts" / "solidity" / f"{contract_name}_abi.json"
            with open(abi_path, 'w') as f:
                json.dump(contract_interface, f, indent=2)
            
            logger.info(f"Contract {contract_name} compiled successfully")
            return contract_interface, contract_bytecode
        
        except Exception as e:
            logger.error(f"Error compiling contract {contract_name}: {str(e)}")
            raise
    
    def get_contract_abi(self, contract_name: str) -> List[Dict[str, Any]]:
        """
        Get ABI for a compiled contract
        
        Args:
            contract_name: Name of the contract
            
        Returns:
            Contract ABI
        """
        try:
            # Check if ABI file exists
            abi_path = Path(__file__).parent / "contracts" / "solidity" / f"{contract_name}_abi.json"
            
            if not abi_path.exists():
                # Try to compile the contract
                contract_interface, _ = self.compile_solidity_contract(contract_name)
                return contract_interface
            
            # Load ABI from file
            with open(abi_path, 'r') as f:
                contract_interface = json.load(f)
            
            return contract_interface
        
        except Exception as e:
            logger.error(f"Error loading ABI for {contract_name}: {str(e)}")
            raise