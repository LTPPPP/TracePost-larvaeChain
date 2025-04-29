from typing import Dict, Any, List, Optional
from web3 import Web3
import json
from pathlib import Path
import os

from app.utils.logger import get_logger

logger = get_logger(__name__)

# ABI for the EventLog contract
EVENT_LOG_ABI = [
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
                "name": "documentHash",
                "type": "string"
            },
            {
                "indexed": False,
                "internalType": "string",
                "name": "documentId",
                "type": "string"
            }
        ],
        "name": "DocumentLogged",
        "type": "event"
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
                "indexed": True,
                "internalType": "string",
                "name": "eventId",
                "type": "string"
            },
            {
                "indexed": False,
                "internalType": "string",
                "name": "eventType",
                "type": "string"
            }
        ],
        "name": "EventLogged",
        "type": "event"
    },
    {
        "inputs": [
            {
                "internalType": "string",
                "name": "",
                "type": "string"
            }
        ],
        "name": "documents",
        "outputs": [
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
                "name": "logger",
                "type": "address"
            }
        ],
        "stateMutability": "view",
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
        "name": "events",
        "outputs": [
            {
                "internalType": "string",
                "name": "shipmentId",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "eventType",
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
                "name": "logger",
                "type": "address"
            }
        ],
        "stateMutability": "view",
        "type": "function"
    },
    {
        "inputs": [
            {
                "internalType": "string",
                "name": "documentHash",
                "type": "string"
            }
        ],
        "name": "getDocument",
        "outputs": [
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
                "name": "eventId",
                "type": "string"
            }
        ],
        "name": "getEvent",
        "outputs": [
            {
                "internalType": "string",
                "name": "shipmentId",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "eventType",
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
                "name": "documentId",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "documentHash",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "metadata",
                "type": "string"
            }
        ],
        "name": "logDocument",
        "outputs": [],
        "stateMutability": "nonpayable",
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
                "name": "eventId",
                "type": "string"
            },
            {
                "internalType": "string",
                "name": "eventType",
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
        "name": "logEvent",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
    }
]

class EventLog:
    """Interface for the EventLog Ethereum smart contract"""
    
    def __init__(self, web3: Web3, address: str):
        """
        Initialize the EventLog contract interface
        
        Args:
            web3: Web3 instance
            address: Contract address
        """
        self.web3 = web3
        self.contract_address = address
        self.contract = web3.eth.contract(address=address, abi=EVENT_LOG_ABI)
        logger.info(f"EventLog contract initialized at {address}")
    
    def build_log_event_transaction(
        self,
        shipment_id: str,
        event_id: str,
        event_type: str,
        data_hash: str,
        metadata: str,
        sender: str
    ) -> Dict[str, Any]:
        """
        Build a transaction to log an event
        
        Args:
            shipment_id: Shipment identifier
            event_id: Unique identifier for the event
            event_type: Type of event
            data_hash: Hash of the event data
            metadata: JSON string with additional metadata
            sender: Ethereum address of the transaction sender
            
        Returns:
            Transaction dictionary
        """
        # Get current gas price with a small bump for faster confirmation
        gas_price = int(self.web3.eth.gas_price * 1.1)
        
        # Build the transaction
        tx = self.contract.functions.logEvent(
            shipment_id,
            event_id,
            event_type,
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
    
    def build_log_document_transaction(
        self,
        document_id: str,
        document_hash: str,
        metadata: str,
        sender: str
    ) -> Dict[str, Any]:
        """
        Build a transaction to log a document
        
        Args:
            document_id: Unique identifier for the document
            document_hash: Hash of the document
            metadata: JSON string with additional metadata
            sender: Ethereum address of the transaction sender
            
        Returns:
            Transaction dictionary
        """
        # Get current gas price with a small bump for faster confirmation
        gas_price = int(self.web3.eth.gas_price * 1.1)
        
        # Build the transaction
        tx = self.contract.functions.logDocument(
            document_id,
            document_hash,
            metadata
        ).build_transaction({
            'from': sender,
            'nonce': self.web3.eth.get_transaction_count(sender),
            'gas': 500000,  # Gas limit
            'gasPrice': gas_price,
            'chainId': self.web3.eth.chain_id
        })
        
        return tx
    
    def get_event(self, event_id: str) -> tuple:
        """
        Get event details from the contract
        
        Args:
            event_id: Event ID
            
        Returns:
            Tuple with event details
        """
        return self.contract.functions.getEvent(event_id).call()
    
    def get_document(self, document_hash: str) -> tuple:
        """
        Get document details from the contract
        
        Args:
            document_hash: Document hash
            
        Returns:
            Tuple with document details
        """
        return self.contract.functions.getDocument(document_hash).call()
    
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
                    # Try to decode as EventLogged
                    try:
                        decoded_log = self.contract.events.EventLogged().process_log(log)
                        decoded_logs.append(decoded_log)
                        continue
                    except:
                        pass
                    
                    # Try to decode as DocumentLogged
                    try:
                        decoded_log = self.contract.events.DocumentLogged().process_log(log)
                        decoded_logs.append(decoded_log)
                    except:
                        pass
            except Exception as e:
                logger.warning(f"Failed to decode log: {str(e)}")
                continue
        
        return decoded_logs