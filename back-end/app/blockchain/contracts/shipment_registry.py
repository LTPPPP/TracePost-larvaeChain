from typing import Dict, Any, List, Optional
from web3 import Web3
import json
from pathlib import Path
import os

from app.utils.logger import get_logger

logger = get_logger(__name__)

# ABI for the ShipmentRegistry contract
SHIPMENT_REGISTRY_ABI = [
    {
        "inputs": [],
        "stateMutability": "nonpayable",
        "type": "constructor"
    },
    {
        "anonymous": False,
        "inputs": [
            {
                "indexed": True,
                "internalType": "string",
                "name": "shipmentId",
                "type": "string"
            },
            {
                "indexed": False,
                "internalType": "string",
                "name": "trackingNumber",
                "type": "string"
            },
            {
                "indexed": False,
                "internalType": "string",
                "name": "dataHash",
                "type": "string"
            }
        ],
        "name": "ShipmentRegistered",
        "type": "event"
    },
    {
        "inputs": [
            {
                "internalType": "string",
                "name": "shipmentId",
                "type": "string"
            }
        ],
        "name": "getShipment",
        "outputs": [
            {
                "internalType": "string",
                "name": "trackingNumber",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "dataHash",
                "type": "string"
            },
            {
                "internalType": "uint256",
                "name": "timestamp",
                "type": "uint256"
            },
            {
                "internalType": "string",
                "name": "metadata",
                "type": "string"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "string",
                "name": "shipmentId",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "trackingNumber",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "dataHash",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "metadata",
                "type": "string"
            }
        ],
        "name": "registerShipment",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "string",
                "name": "",
                "type": "string"
            }
        ],
        "name": "shipments",
        "outputs": [
            {
                "internalType": "string",
                "name": "trackingNumber",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "dataHash",
                "type": "string"
            },
            {
                "internalType": "uint256",
                "name": "timestamp",
                "type": "uint256"
            },
            {
                "internalType": "string",
                "name": "metadata",
                "type": "string"
            },
            {
                "internalType": "address",
                "name": "registrar",
                "type": "address"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    }
]

class ShipmentRegistry:
    """Interface for the ShipmentRegistry Ethereum smart contract"""
    
    def __init__(self, web3: Web3, address: str):
        """
        Initialize the ShipmentRegistry contract interface
        
        Args:
            web3: Web3 instance
            address: Contract address
        """
        self.web3 = web3
        self.contract_address = address
        self.contract = web3.eth.contract(address=address, abi=SHIPMENT_REGISTRY_ABI)
        logger.info(f"ShipmentRegistry contract initialized at {address}")
    
    def build_register_shipment_transaction(
        self,
        shipment_id: str,
        tracking_number: str,
        data_hash: str,
        metadata: str,
        sender: str
    ) -> Dict[str, Any]:
        """
        Build a transaction to register a shipment
        
        Args:
            shipment_id: Unique identifier for the shipment
            tracking_number: Shipment tracking number
            data_hash: Hash of the shipment data
            metadata: JSON string with additional metadata
            sender: Ethereum address of the transaction sender
            
        Returns:
            Transaction dictionary
        """
        # Get current gas price with a small bump for faster confirmation
        gas_price = int(self.web3.eth.gas_price * 1.1)
        
        # Build the transaction
        tx = self.contract.functions.registerShipment(
            shipment_id,
            tracking_number,
            data_hash,
            metadata
        ).build_transaction({
            'from': sender,
            'nonce': self.web3.eth.get_transaction_count(sender),
            'gas': 500000,  # Gas limit
            'gasPrice': gas_price,
            'chainId': self.web3.eth.chain_id
        })
        
        return tx
    
    def get_shipment(self, shipment_id: str) -> tuple:
        """
        Get shipment details from the contract
        
        Args:
            shipment_id: Shipment ID
            
        Returns:
            Tuple with shipment details
        """
        return self.contract.functions.getShipment(shipment_id).call()
    
    def decode_logs(self, logs: List[Dict[str, Any]]) -> List[Any]:
        """
        Decode event logs from a transaction receipt
        
        Args:
            logs: List of log dictionaries from transaction receipt
            
        Returns:
            List of decoded events
        """
        decoded_logs = []
        for log in logs:
            try:
                if log['address'].lower() == self.contract_address.lower():
                    decoded_log = self.contract.events.ShipmentRegistered().process_log(log)
                    decoded_logs.append(decoded_log)
            except Exception as e:
                logger.warning(f"Failed to decode log: {str(e)}")
                continue
        
        return decoded_logs