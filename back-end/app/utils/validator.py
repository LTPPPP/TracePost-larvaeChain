import re
import uuid
from datetime import datetime
from typing import Optional, Any

from app.utils.logger import get_logger

logger = get_logger(__name__)

def validate_uuid(value: Any) -> Optional[uuid.UUID]:
    """
    Validate that a value is a valid UUID
    
    Args:
        value: Value to validate
        
    Returns:
        UUID object if valid, None otherwise
    """
    if not value:
        return None
        
    try:
        if isinstance(value, str):
            return uuid.UUID(value)
        elif isinstance(value, uuid.UUID):
            return value
        else:
            return None
    except (ValueError, AttributeError, TypeError):
        logger.debug(f"Invalid UUID: {value}")
        return None

def validate_tracking_number(tracking_number: str) -> bool:
    """
    Validate tracking number format
    
    Args:
        tracking_number: Tracking number to validate
        
    Returns:
        True if valid, False otherwise
    """
    if not tracking_number:
        return False
        
    # Tracking number format: 3 letters + 9-12 digits
    pattern = r'^[A-Z]{3}[0-9]{9,12}$'
    return bool(re.match(pattern, tracking_number))

def validate_blockchain_hash(hash_str: str) -> bool:
    """
    Validate blockchain transaction hash
    
    Args:
        hash_str: Hash to validate
        
    Returns:
        True if valid, False otherwise
    """
    if not hash_str:
        return False
        
    # Ethereum style tx hash: 0x + 64 hex chars
    eth_pattern = r'^0x[a-fA-F0-9]{64}$'
    
    # VN Chain or other hash format (more permissive)
    other_pattern = r'^[a-zA-Z0-9]{40,128}$'
    
    return bool(re.match(eth_pattern, hash_str) or re.match(other_pattern, hash_str))

def validate_date_range(start_date: Optional[datetime], end_date: Optional[datetime]) -> bool:
    """
    Validate that start_date is before end_date
    
    Args:
        start_date: Start date
        end_date: End date
        
    Returns:
        True if valid range, False otherwise
    """
    # If either date is None, consider valid
    if start_date is None or end_date is None:
        return True
        
    return start_date <= end_date

def validate_coordinates(latitude: Optional[float], longitude: Optional[float]) -> bool:
    """
    Validate geographical coordinates
    
    Args:
        latitude: Latitude to validate (-90 to 90)
        longitude: Longitude to validate (-180 to 180)
        
    Returns:
        True if valid coordinates, False otherwise
    """
    if latitude is None or longitude is None:
        return False
        
    return -90 <= latitude <= 90 and -180 <= longitude <= 180

def generate_tracking_number(prefix: str = "VCN") -> str:
    """
    Generate a unique tracking number
    
    Args:
        prefix: Three-letter prefix for the tracking number
        
    Returns:
        Tracking number in the format PREFIX + TIMESTAMP + RANDOM
    """
    # Ensure prefix is exactly 3 uppercase letters
    prefix = prefix[:3].upper()
    if len(prefix) < 3:
        prefix = prefix.ljust(3, 'X')
        
    # Get current timestamp (milliseconds since epoch)
    timestamp = int(datetime.now().timestamp() * 1000)
    
    # Add a random component (last 4 digits of a UUID)
    random_part = str(uuid.uuid4().int)[-4:]
    
    # Combine parts to create a 15-digit tracking number
    tracking_number = f"{prefix}{timestamp}{random_part}"
    
    # Ensure it's not longer than the expected format
    return tracking_number[:15]