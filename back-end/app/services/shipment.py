from typing import List, Optional, Dict, Any, Tuple
from uuid import UUID
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession

from app.db.repositories import (
    shipment_repository,
    shipment_item_repository,
    document_repository
)
from app.schemas.shipment import (
    ShipmentCreate, 
    ShipmentUpdate,
    ShipmentItemCreate,
    ShipmentItemUpdate,
    DocumentCreate,
    ShipmentWithItems,
    ShipmentWithEvents,
    ShipmentWithItemsAndEvents
)
from app.models.shipment import Shipment, ShipmentItem, Document
from app.models.user import User
from app.utils.validator import validate_tracking_number, generate_tracking_number
from app.core.exceptions import (
    ResourceNotFoundError,
    ResourceAlreadyExistsError,
    ValidationError
)
from app.utils.logger import get_logger

logger = get_logger(__name__)

async def create_shipment(
    db: AsyncSession, 
    shipment_in: ShipmentCreate,
    user: User
) -> Shipment:
    """
    Create a new shipment
    
    Args:
        db: Database session
        shipment_in: Shipment data
        user: Current user
        
    Returns:
        Created shipment
        
    Raises:
        ResourceAlreadyExistsError: If shipment with same tracking number exists
        ValidationError: If tracking number is invalid
    """
    # Generate tracking number if not provided
    if not shipment_in.tracking_number:
        shipment_in.tracking_number = generate_tracking_number()
    else:
        # Validate tracking number format
        if not validate_tracking_number(shipment_in.tracking_number):
            logger.warning(f"Invalid tracking number format: {shipment_in.tracking_number}")
            raise ValidationError(detail="Invalid tracking number format")
    
    # Check if a shipment with same tracking number already exists
    existing = await shipment_repository.get_by_tracking_number(db, shipment_in.tracking_number)
    if existing:
        logger.warning(f"Shipment with tracking number {shipment_in.tracking_number} already exists")
        raise ResourceAlreadyExistsError(detail="Shipment with this tracking number already exists")
    
    # Create shipment
    shipment = await shipment_repository.create(db, obj_in=shipment_in)
    logger.info(f"Shipment created: {shipment.tracking_number} (ID: {shipment.id})")
    
    return shipment

async def update_shipment(
    db: AsyncSession, 
    shipment_id: UUID, 
    shipment_in: ShipmentUpdate,
    user: User
) -> Shipment:
    """
    Update a shipment
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        shipment_in: New shipment data
        user: Current user
        
    Returns:
        Updated shipment
        
    Raises:
        ResourceNotFoundError: If shipment not found
        ValidationError: If tracking number is invalid
    """
    # Get the shipment
    shipment = await shipment_repository.get(db, id=shipment_id)
    if not shipment:
        logger.warning(f"Shipment update failed: Shipment {shipment_id} not found")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Check if the user has permission to update this shipment
    if user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to update shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Validate tracking number if provided
    if shipment_in.tracking_number and not validate_tracking_number(shipment_in.tracking_number):
        logger.warning(f"Invalid tracking number format: {shipment_in.tracking_number}")
        raise ValidationError(detail="Invalid tracking number format")
    
    # Update shipment
    shipment = await shipment_repository.update(db, db_obj=shipment, obj_in=shipment_in)
    logger.info(f"Shipment updated: {shipment.tracking_number} (ID: {shipment.id})")
    
    return shipment

async def get_shipment(
    db: AsyncSession, 
    shipment_id: UUID,
    user: Optional[User] = None
) -> Shipment:
    """
    Get a shipment by ID
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        user: Current user
        
    Returns:
        Shipment if found
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    shipment = await shipment_repository.get(db, id=shipment_id)
    
    if not shipment:
        logger.warning(f"Shipment not found: {shipment_id}")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Check if the user has permission to view this shipment
    if user and user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to view shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    return shipment

async def get_shipment_by_tracking(
    db: AsyncSession, 
    tracking_number: str,
    user: Optional[User] = None
) -> Shipment:
    """
    Get a shipment by tracking number
    
    Args:
        db: Database session
        tracking_number: Shipment tracking number
        user: Current user
        
    Returns:
        Shipment if found
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    shipment = await shipment_repository.get_by_tracking_number(db, tracking_number)
    
    if not shipment:
        logger.warning(f"Shipment not found with tracking number: {tracking_number}")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Check if the user has permission to view this shipment
    if user and user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to view shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    return shipment

async def get_shipment_with_details(
    db: AsyncSession, 
    shipment_id: UUID,
    user: Optional[User] = None
) -> ShipmentWithItemsAndEvents:
    """
    Get a shipment with items and events
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        user: Current user
        
    Returns:
        Shipment with items and events
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    shipment = await shipment_repository.get_with_details(db, id=shipment_id)
    
    if not shipment:
        logger.warning(f"Shipment not found: {shipment_id}")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Check if the user has permission to view this shipment
    if user and user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to view shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    return shipment

async def list_shipments(
    db: AsyncSession, 
    user: User,
    skip: int = 0, 
    limit: int = 100,
    status: Optional[str] = None
) -> List[Shipment]:
    """
    List shipments for an organization
    
    Args:
        db: Database session
        user: Current user
        skip: Number of records to skip
        limit: Maximum number of records to return
        status: Filter by status
        
    Returns:
        List of shipments
    """
    # If user is admin and no organization is specified, return all shipments
    if user.role == "admin":
        # For admins, return all shipments regardless of organization
        return await shipment_repository.get_multi(
            db, 
            skip=skip, 
            limit=limit,
            status=status
        )
    else:
        # For regular users, only return shipments from their organization
        return await shipment_repository.get_by_organization(
            db, 
            organization_id=user.organization_id,
            skip=skip, 
            limit=limit,
            status=status
        )

async def search_shipments(
    db: AsyncSession, 
    user: User,
    query: Optional[str] = None,
    status: Optional[str] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    is_international: Optional[bool] = None,
    skip: int = 0, 
    limit: int = 100
) -> List[Shipment]:
    """
    Search shipments with filtering
    
    Args:
        db: Database session
        user: Current user
        query: Search query for title or tracking number
        status: Filter by status
        start_date: Start date for created_at
        end_date: End date for created_at
        is_international: Filter by international status
        skip: Number of records to skip
        limit: Maximum number of records to return
        
    Returns:
        List of shipments matching criteria
    """
    # If user is admin, search all shipments
    organization_id = None if user.role == "admin" else user.organization_id
    
    return await shipment_repository.search(
        db,
        organization_id=organization_id,
        query=query,
        status=status,
        start_date=start_date,
        end_date=end_date,
        is_international=is_international,
        skip=skip,
        limit=limit
    )

async def get_shipment_status_counts(
    db: AsyncSession, 
    user: User
) -> Dict[str, int]:
    """
    Get counts of shipments by status
    
    Args:
        db: Database session
        user: Current user
        
    Returns:
        Dictionary of status -> count
    """
    # If user is admin, get counts for all shipments
    organization_id = None if user.role == "admin" else user.organization_id
    
    return await shipment_repository.get_shipment_counts(
        db,
        organization_id=organization_id
    )

async def add_shipment_item(
    db: AsyncSession, 
    item_in: ShipmentItemCreate,
    user: User
) -> ShipmentItem:
    """
    Add an item to a shipment
    
    Args:
        db: Database session
        item_in: Item data
        user: Current user
        
    Returns:
        Created shipment item
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    # Check if shipment exists and user has access
    shipment = await get_shipment(db, item_in.shipment_id, user)
    
    # Create item
    item = await shipment_item_repository.create(db, obj_in=item_in)
    logger.info(f"Item added to shipment {shipment.tracking_number}: {item.name} (ID: {item.id})")
    
    return item

async def update_shipment_item(
    db: AsyncSession, 
    item_id: UUID,
    item_in: ShipmentItemUpdate,
    user: User
) -> ShipmentItem:
    """
    Update a shipment item
    
    Args:
        db: Database session
        item_id: Item ID
        item_in: New item data
        user: Current user
        
    Returns:
        Updated shipment item
        
    Raises:
        ResourceNotFoundError: If item not found
    """
    # Get the item
    item = await shipment_item_repository.get(db, id=item_id)
    if not item:
        logger.warning(f"Item update failed: Item {item_id} not found")
        raise ResourceNotFoundError(detail="Item not found")
    
    # Check if user has access to the shipment
    await get_shipment(db, item.shipment_id, user)
    
    # Update item
    item = await shipment_item_repository.update(db, db_obj=item, obj_in=item_in)
    logger.info(f"Item updated: {item.name} (ID: {item.id})")
    
    return item

async def delete_shipment_item(
    db: AsyncSession, 
    item_id: UUID,
    user: User
) -> ShipmentItem:
    """
    Delete a shipment item
    
    Args:
        db: Database session
        item_id: Item ID
        user: Current user
        
    Returns:
        Deleted shipment item
        
    Raises:
        ResourceNotFoundError: If item not found
    """
    # Get the item
    item = await shipment_item_repository.get(db, id=item_id)
    if not item:
        logger.warning(f"Item deletion failed: Item {item_id} not found")
        raise ResourceNotFoundError(detail="Item not found")
    
    # Check if user has access to the shipment
    await get_shipment(db, item.shipment_id, user)
    
    # Delete item
    deleted_item = await shipment_item_repository.delete(db, id=item_id)
    logger.info(f"Item deleted: {deleted_item.name} (ID: {deleted_item.id})")
    
    return deleted_item

async def get_shipment_items(
    db: AsyncSession, 
    shipment_id: UUID,
    user: User
) -> List[ShipmentItem]:
    """
    Get items for a shipment
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        user: Current user
        
    Returns:
        List of shipment items
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    # Check if shipment exists and user has access
    await get_shipment(db, shipment_id, user)
    
    # Get items
    return await shipment_item_repository.get_by_shipment(db, shipment_id=shipment_id)

async def update_blockchain_status(
    db: AsyncSession,
    shipment_id: UUID,
    tx_hash: str,
    network: str,
    status: str = "pending"
) -> Shipment:
    """
    Update blockchain status for a shipment
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        tx_hash: Blockchain transaction hash
        network: Blockchain network
        status: Transaction status
        
    Returns:
        Updated shipment
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    shipment = await shipment_repository.update_blockchain_status(
        db,
        shipment_id=shipment_id,
        tx_hash=tx_hash,
        network=network,
        status=status
    )
    
    if not shipment:
        logger.warning(f"Blockchain status update failed: Shipment {shipment_id} not found")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    logger.info(f"Blockchain status updated for shipment {shipment.tracking_number}: {status}")
    return shipment