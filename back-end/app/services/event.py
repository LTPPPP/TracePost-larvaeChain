from typing import List, Optional, Dict, Any, Tuple
from uuid import UUID
from datetime import datetime

from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends

from app.db.database import get_db
from app.db.repositories import (
    event_repository,
    shipment_repository
)
from app.schemas.event import (
    EventCreate,
    EventUpdate,
    Timeline,
    TimelineEvent,
    ShipmentEventStats
)
from app.models.event import ShipmentEvent
from app.models.shipment import Shipment
from app.models.user import User
from app.core.exceptions import (
    ResourceNotFoundError,
    ValidationError
)
from app.utils.validator import validate_coordinates
from app.utils.logger import get_logger

logger = get_logger(__name__)

async def create_event(
    db: AsyncSession,
    event_in: EventCreate,
    user: User
) -> ShipmentEvent:
    """
    Create a new shipment event
    
    Args:
        db: Database session
        event_in: Event data
        user: Current user
        
    Returns:
        Created event
        
    Raises:
        ResourceNotFoundError: If shipment not found
        ValidationError: If coordinates are invalid
    """
    # Check if shipment exists
    shipment = await shipment_repository.get(db, id=event_in.shipment_id)
    if not shipment:
        logger.warning(f"Event creation failed: Shipment {event_in.shipment_id} not found")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Check if user has access to the shipment
    if user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to create event for shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Validate coordinates if provided
    if event_in.latitude is not None and event_in.longitude is not None:
        if not validate_coordinates(event_in.latitude, event_in.longitude):
            logger.warning(f"Invalid coordinates: {event_in.latitude}, {event_in.longitude}")
            raise ValidationError(detail="Invalid coordinates")
    
    # Set the user and organization ID
    event_data = event_in.dict()
    event_data["user_id"] = user.id
    event_data["organization_id"] = user.organization_id
    
    # Create event
    event = await event_repository.create(db, obj_in=event_data)
    logger.info(f"Event created: {event.event_type} for shipment {shipment.tracking_number} (ID: {event.id})")
    
    # Update shipment status based on event type if applicable
    if event.event_type == "pickup":
        shipment.status = "picked_up"
        db.add(shipment)
        await db.commit()
    elif event.event_type == "delivery":
        shipment.status = "delivered"
        shipment.actual_delivery = event.timestamp
        db.add(shipment)
        await db.commit()
    elif event.event_type == "customs_check":
        shipment.status = "customs"
        db.add(shipment)
        await db.commit()
    elif event.event_type == "transit":
        shipment.status = "in_transit"
        db.add(shipment)
        await db.commit()
    elif event.event_type == "delay":
        shipment.status = "delayed"
        db.add(shipment)
        await db.commit()
    
    # TODO: Process IoT data alerts if applicable
    
    return event

async def update_event(
    db: AsyncSession,
    event_id: UUID,
    event_in: EventUpdate,
    user: User
) -> ShipmentEvent:
    """
    Update a shipment event
    
    Args:
        db: Database session
        event_id: Event ID
        event_in: New event data
        user: Current user
        
    Returns:
        Updated event
        
    Raises:
        ResourceNotFoundError: If event not found
        ValidationError: If coordinates are invalid
    """
    # Get the event
    event = await event_repository.get(db, id=event_id)
    if not event:
        logger.warning(f"Event update failed: Event {event_id} not found")
        raise ResourceNotFoundError(detail="Event not found")
    
    # Check if user has access to the shipment
    shipment = await shipment_repository.get(db, id=event.shipment_id)
    if not shipment:
        logger.warning(f"Event update failed: Associated shipment {event.shipment_id} not found")
        raise ResourceNotFoundError(detail="Event not found")
    
    if user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to update event for shipment from another organization")
        raise ResourceNotFoundError(detail="Event not found")
    
    # Validate coordinates if provided
    if event_in.latitude is not None and event_in.longitude is not None:
        if not validate_coordinates(event_in.latitude, event_in.longitude):
            logger.warning(f"Invalid coordinates: {event_in.latitude}, {event_in.longitude}")
            raise ValidationError(detail="Invalid coordinates")
    
    # Update event
    event = await event_repository.update(db, db_obj=event, obj_in=event_in)
    logger.info(f"Event updated: {event.event_type} (ID: {event.id})")
    
    return event

async def get_event(
    db: AsyncSession,
    event_id: UUID,
    user: Optional[User] = None
) -> ShipmentEvent:
    """
    Get an event by ID
    
    Args:
        db: Database session
        event_id: Event ID
        user: Current user
        
    Returns:
        Event if found
        
    Raises:
        ResourceNotFoundError: If event not found
    """
    event = await event_repository.get(db, id=event_id)
    
    if not event:
        logger.warning(f"Event not found: {event_id}")
        raise ResourceNotFoundError(detail="Event not found")
    
    # Check if user has access to the shipment
    if user:
        shipment = await shipment_repository.get(db, id=event.shipment_id)
        if not shipment:
            logger.warning(f"Event retrieval failed: Associated shipment {event.shipment_id} not found")
            raise ResourceNotFoundError(detail="Event not found")
        
        if user.role != "admin" and shipment.organization_id != user.organization_id:
            logger.warning(f"User {user.id} attempted to access event for shipment from another organization")
            raise ResourceNotFoundError(detail="Event not found")
    
    return event

async def get_shipment_events(
    db: AsyncSession,
    shipment_id: UUID,
    user: User,
    skip: int = 0,
    limit: int = 100,
    event_type: Optional[str] = None,
    source: Optional[str] = None
) -> List[ShipmentEvent]:
    """
    Get events for a shipment
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        user: Current user
        skip: Number of records to skip
        limit: Maximum number of records to return
        event_type: Filter by event type
        source: Filter by event source
        
    Returns:
        List of events
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    # Check if shipment exists and user has access
    shipment = await shipment_repository.get(db, id=shipment_id)
    if not shipment:
        logger.warning(f"Events retrieval failed: Shipment {shipment_id} not found")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    if user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to access events for shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Get events
    return await event_repository.get_by_shipment(
        db,
        shipment_id=shipment_id,
        skip=skip,
        limit=limit,
        event_type=event_type,
        source=source
    )

async def get_event_timeline(
    db: AsyncSession,
    shipment_id: UUID,
    user: Optional[User] = None
) -> Timeline:
    """
    Get timeline of events for a shipment
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        user: Current user
        
    Returns:
        Timeline object
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    # Check if shipment exists and user has access
    shipment = await shipment_repository.get(db, id=shipment_id)
    if not shipment:
        logger.warning(f"Timeline retrieval failed: Shipment {shipment_id} not found")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    if user and user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to access timeline for shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Get events ordered by timestamp
    events, total = await event_repository.get_timeline(db, shipment_id=shipment_id)
    
    # Map events to timeline format
    timeline_events = []
    for event in events:
        timeline_events.append(
            TimelineEvent(
                id=event.id,
                event_type=event.event_type,
                location=event.location,
                timestamp=event.timestamp,
                description=event.description,
                source=event.source,
                verified=bool(event.verified_by),
                blockchain_verified=(event.blockchain_status == "confirmed")
            )
        )
    
    return Timeline(
        shipment_id=shipment_id,
        tracking_number=shipment.tracking_number,
        events=timeline_events,
        total_events=total
    )

async def get_event_statistics(
    db: AsyncSession,
    shipment_id: UUID,
    user: Optional[User] = None
) -> ShipmentEventStats:
    """
    Get statistics about events for a shipment
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        user: Current user
        
    Returns:
        ShipmentEventStats object
        
    Raises:
        ResourceNotFoundError: If shipment not found
    """
    # Check if shipment exists and user has access
    shipment = await shipment_repository.get(db, id=shipment_id)
    if not shipment:
        logger.warning(f"Event statistics retrieval failed: Shipment {shipment_id} not found")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    if user and user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to access event statistics for shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Get event statistics
    stats = await event_repository.get_event_stats(db, shipment_id=shipment_id)
    
    return ShipmentEventStats(
        shipment_id=shipment_id,
        tracking_number=shipment.tracking_number,
        total_events=stats["total_events"],
        events_by_type=stats["events_by_type"],
        events_by_source=stats["events_by_source"],
        last_event=TimelineEvent.from_orm(stats["last_event"]) if stats["last_event"] else None,
        blockchain_verified_count=stats["blockchain_verified_count"],
        temperature_stats=stats["temperature_stats"],
        humidity_stats=stats["humidity_stats"]
    )

async def get_recent_events(
    db: AsyncSession,
    user: User,
    hours: int = 24,
    limit: int = 10
) -> List[ShipmentEvent]:
    """
    Get recent events
    
    Args:
        db: Database session
        user: Current user
        hours: Number of hours to look back
        limit: Maximum number of records to return
        
    Returns:
        List of recent events
    """
    # If user is admin, get events for all organizations
    organization_id = None if user.role == "admin" else user.organization_id
    
    return await event_repository.get_recent_events(
        db,
        hours=hours,
        organization_id=organization_id,
        limit=limit
    )

async def update_blockchain_status(
    db: AsyncSession,
    event_id: UUID,
    tx_hash: str,
    network: str,
    status: str = "pending"
) -> ShipmentEvent:
    """
    Update blockchain status for an event
    
    Args:
        db: Database session
        event_id: Event ID
        tx_hash: Blockchain transaction hash
        network: Blockchain network
        status: Transaction status
        
    Returns:
        Updated event
        
    Raises:
        ResourceNotFoundError: If event not found
    """
    event = await event_repository.update_blockchain_status(
        db,
        event_id=event_id,
        tx_hash=tx_hash,
        network=network,
        status=status
    )
    
    if not event:
        logger.warning(f"Blockchain status update failed: Event {event_id} not found")
        raise ResourceNotFoundError(detail="Event not found")
    
    logger.info(f"Blockchain status updated for event {event.id}: {status}")
    return event