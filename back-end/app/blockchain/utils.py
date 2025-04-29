# utils.py

from typing import Optional
import re
import hashlib
import json
from eth_utils import is_hex_address, to_checksum_address
from app.utils.validator import validate_blockchain_hash

def normalize_address(address: str) -> Optional[str]:
    """
    Normalize an Ethereum address to checksum format
    
    Args:
        address: Ethereum address
        
    Returns:
        Normalized address or None if invalid
    """
    if not address:
        return None
        
    # Remove 0x prefix if present for validation
    clean_address = address[2:] if address.startswith('0x') else address
    
    # Check if it's a valid hex string of the right length
    if not re.match(r'^[0-9a-fA-F]{40}$', clean_address):
        return None
    
    # Add 0x prefix back if it was missing
    if not address.startswith('0x'):
        address = '0x' + address
    
    try:
        # Convert to checksum address
        return to_checksum_address(address)
    except:
        return None

def is_valid_eth_address(address: str) -> bool:
    """
    Check if an address is a valid Ethereum address
    
    Args:
        address: Ethereum address to validate
        
    Returns:
        True if valid, False otherwise
    """
    if not address:
        return False
        
    try:
        return is_hex_address(address)
    except:
        return False

def hash_data(data: dict) -> str:
    """
    Create a deterministic hash of dictionary data
    
    Args:
        data: Dictionary to hash
        
    Returns:
        Hex string of the hash
    """
    # Sort keys for deterministic serialization
    serialized = json.dumps(data, sort_keys=True).encode()
    return hashlib.sha256(serialized).hexdigest()

def select_blockchain_client(network: str):
    """
    Get the appropriate blockchain client based on network name
    
    Args:
        network: Network name (ethereum, substrate, vietnamchain)
        
    Returns:
        Blockchain client or None if not found/enabled
    """
    from app.blockchain import ethereum_client, substrate_client, vietnamchain_client
    
    if network.lower() == "ethereum":
        return ethereum_client
    elif network.lower() == "substrate":
        return substrate_client
    elif network.lower() == "vietnamchain":
        return vietnamchain_client
    else:
        return None