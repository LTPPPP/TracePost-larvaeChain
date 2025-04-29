from typing import Optional, List, Dict, Any, Tuple
from uuid import UUID
from datetime import datetime, timedelta

from sqlalchemy import select, func, and_, or_, desc, case
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import joinedload

from app.db.repositories.base import BaseRepository
from app.models.event import ShipmentEvent, ShipmentAlert, AuditLog
from app.schemas.event import EventCreate, EventUpdate
from app.utils.logger import get_logger

logger = get_logger(__name__)


class EventRepository(BaseRepository[ShipmentEvent, EventCreate, EventUpdate]):
    """Repository for ShipmentEvent operations"""
    
    def __init__(self):
        super().__init__(ShipmentEvent)
    
    async def get_by_shipment(
        self,
        db: AsyncSession,
        *,
        shipment_id: UUID,
        skip: int = 0,
        limit: int = 100,
        event_type: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        source: Optional[str] = None,
    ) -> List[ShipmentEvent]:
        """
        Get events for a shipment with optional filtering
        
        Args:
            db: Database session
            shipment_id: Shipment ID
            skip: Number of records to skip
            limit: Maximum number of records to return
            event_type: Filter by event type
            start_date: Filter by timestamp start
            end_date: Filter by timestamp end
            source: Filter by event source
            
        Returns:
            List of events
        """
        filters = [ShipmentEvent.shipment_id == shipment_id]
        
        if event_type:
            filters.append(ShipmentEvent.event_type == event_type)
        
        if start_date:
            filters.append(ShipmentEvent.timestamp >= start_date)
        
        if end_date:
            filters.append(ShipmentEvent.timestamp <= end_date)
        
        if source:
            filters.append(ShipmentEvent.source == source)
        
        query = select(ShipmentEvent).where(and_(*filters))
        query = query.order_by(desc(ShipmentEvent.timestamp)).offset(skip).limit(limit)
        
        result = await db.execute(query)
        return result.scalars().all()
    
    async def get_timeline(
        self,
        db: AsyncSession,
        *,
        shipment_id: UUID,
    ) -> Tuple[List[ShipmentEvent], int]:
        """
        Get timeline events for a shipment ordered by timestamp
        
        Args:
            db: Database session
            shipment_id: Shipment ID
            
        Returns:
            Tuple of (events list, total count)
        """
        # First get the count
        count_query = select(func.count(ShipmentEvent.id)).where(
            ShipmentEvent.shipment_id == shipment_id
        )
        count_result = await db.execute(count_query)
        total_count = count_result.scalar()
        
        # Then get the events
        query = select(ShipmentEvent).where(
            ShipmentEvent.shipment_id == shipment_id
        ).order_by(ShipmentEvent.timestamp)
        
        result = await db.execute(query)
        events = result.scalars().all()
        
        return events, total_count
    
    async def get_event_stats(
        self,
        db: AsyncSession,
        *,
        shipment_id: UUID,
    ) -> Dict[str, Any]:
        """
        Get event statistics for a shipment
        
        Args:
            db: Database session
            shipment_id: Shipment ID
            
        Returns:
            Dictionary with event statistics
        """
        # Get counts by event type
        type_query = select(
            ShipmentEvent.event_type,
            func.count(ShipmentEvent.id)
        ).where(
            ShipmentEvent.shipment_id == shipment_id
        ).group_by(ShipmentEvent.event_type)
        
        type_result = await db.execute(type_query)
        type_counts = {event_type: count for event_type, count in type_result.all()}
        
        # Get counts by source
        source_query = select(
            ShipmentEvent.source,
            func.count(ShipmentEvent.id)
        ).where(
            ShipmentEvent.shipment_id == shipment_id
        ).group_by(ShipmentEvent.source)
        
        source_result = await db.execute(source_query)
        source_counts = {source: count for source, count in source_result.all()}
        
        # Get blockchain verified count
        blockchain_query = select(func.count(ShipmentEvent.id)).where(
            ShipmentEvent.shipment_id == shipment_id,
            ShipmentEvent.blockchain_tx_hash.isnot(None),
            ShipmentEvent.blockchain_status == "confirmed"
        )
        
        blockchain_result = await db.execute(blockchain_query)
        blockchain_verified_count = blockchain_result.scalar() or 0
        
        # Get latest event
        latest_query = select(ShipmentEvent).where(
            ShipmentEvent.shipment_id == shipment_id
        ).order_by(desc(ShipmentEvent.timestamp)).limit(1)
        
        latest_result = await db.execute(latest_query)
        last_event = latest_result.scalars().first()
        
        # Get temperature stats if available
        temp_query = select(
            func.min(ShipmentEvent.temperature),
            func.max(ShipmentEvent.temperature),
            func.avg(ShipmentEvent.temperature)
        ).where(
            ShipmentEvent.shipment_id == shipment_id,
            ShipmentEvent.temperature.isnot(None)
        )
        
        temp_result = await db.execute(temp_query)
        temp_min, temp_max, temp_avg = temp_result.one()
        
        temperature_stats = None
        if temp_min is not None:
            temperature_stats = {
                "min": temp_min,
                "max": temp_max,
                "avg": temp_avg
            }
        
        # Get humidity stats if available
        humidity_query = select(
            func.min(ShipmentEvent.humidity),
            func.max(ShipmentEvent.humidity),
            func.avg(ShipmentEvent.humidity)
        ).where(
            ShipmentEvent.shipment_id == shipment_id,
            ShipmentEvent.humidity.isnot(None)
        )
        
        humidity_result = await db.execute(humidity_query)
        humidity_min, humidity_max, humidity_avg = humidity_result.one()
        
        humidity_stats = None
        if humidity_min is not None:
            humidity_stats = {
                "min": humidity_min,
                "max": humidity_max,
                "avg": humidity_avg
            }
        
        # Get total count
        total_query = select(func.count(ShipmentEvent.id)).where(
            ShipmentEvent.shipment_id == shipment_id
        )
        
        total_result = await db.execute(total_query)
        total_events = total_result.scalar() or 0
        
        return {
            "total_events": total_events,
            "events_by_type": type_counts,
            "events_by_source": source_counts,
            "last_event": last_event,
            "blockchain_verified_count": blockchain_verified_count,
            "temperature_stats": temperature_stats,
            "humidity_stats": humidity_stats
        }
    
    async def update_blockchain_status(
        self,
        db: AsyncSession,
        *,
        event_id: UUID,
        tx_hash: str,
        network: str,
        status: str = "pending"
    ) -> Optional[ShipmentEvent]:
        """
        Update blockchain status for an event
        
        Args:
            db: Database session
            event_id: Event ID
            tx_hash: Blockchain transaction hash
            network: Blockchain network
            status: Transaction status
            
        Returns:
            Updated event or None if not found
        """
        event = await self.get(db, id=event_id)
        if not event:
            return None
        
        event.blockchain_tx_hash = tx_hash
        event.blockchain_network = network
        event.blockchain_timestamp = datetime.utcnow()
        event.blockchain_status = status
        
        db.add(event)
        await db.commit()
        await db.refresh(event)
        return event
    
    async def get_events_by_date_range(
        self,
        db: AsyncSession,
        *,
        start_date: datetime,
        end_date: datetime,
        organization_id: Optional[UUID] = None,
        event_type: Optional[str] = None,
        skip: int = 0,
        limit: int = 100
    ) -> List[ShipmentEvent]:
        """
        Get events within a date range with optional filtering
        
        Args:
            db: Database session
            start_date: Start date
            end_date: End date
            organization_id: Organization ID to filter by
            event_type: Event type to filter by
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of events
        """
        filters = [
            ShipmentEvent.timestamp >= start_date,
            ShipmentEvent.timestamp <= end_date
        ]
        
        if organization_id:
            filters.append(ShipmentEvent.organization_id == organization_id)
        
        if event_type:
            filters.append(ShipmentEvent.event_type == event_type)
        
        query = select(ShipmentEvent).where(and_(*filters))
        query = query.order_by(desc(ShipmentEvent.timestamp)).offset(skip).limit(limit)
        
        result = await db.execute(query)
        return result.scalars().all()
    
    async def get_recent_events(
        self,
        db: AsyncSession,
        *,
        hours: int = 24,
        organization_id: Optional[UUID] = None,
        limit: int = 10
    ) -> List[ShipmentEvent]:
        """
        Get recent events
        
        Args:
            db: Database session
            hours: Number of hours to look back
            organization_id: Organization ID to filter by
            limit: Maximum number of records to return
            
        Returns:
            List of recent events
        """
        since = datetime.utcnow() - timedelta(hours=hours)
        
        filters = [ShipmentEvent.timestamp >= since]
        
        if organization_id:
            filters.append(ShipmentEvent.organization_id == organization_id)
        
        query = select(ShipmentEvent).where(and_(*filters))
        query = query.order_by(desc(ShipmentEvent.timestamp)).limit(limit)
        
        result = await db.execute(query)
        return result.scalars().all()


class ShipmentAlertRepository(BaseRepository):
    """Repository for ShipmentAlert operations"""
    
    def __init__(self):
        super().__init__(ShipmentAlert)
    
    async def get_by_shipment(
        self,
        db: AsyncSession,
        *,
        shipment_id: UUID,
        resolved: Optional[bool] = None,
        skip: int = 0,
        limit: int = 100
    ) -> List[ShipmentAlert]:
        """
        Get alerts for a shipment
        
        Args:
            db: Database session
            shipment_id: Shipment ID
            resolved: Filter by resolved status
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of alerts
        """
        filters = [ShipmentAlert.shipment_id == shipment_id]
        
        if resolved is not None:
            filters.append(ShipmentAlert.resolved == resolved)
        
        query = select(ShipmentAlert).where(and_(*filters))
        query = query.order_by(desc(ShipmentAlert.created_at)).offset(skip).limit(limit)
        
        result = await db.execute(query)
        return result.scalars().all()


class AuditLogRepository(BaseRepository):
    """Repository for AuditLog operations"""
    
    def __init__(self):
        super().__init__(AuditLog)
    
    async def log_api_call(
        self,
        db: AsyncSession,
        *,
        request_method: str,
        request_path: str,
        request_id: Optional[str] = None,
        request_ip: Optional[str] = None,
        request_user_agent: Optional[str] = None,
        response_status: Optional[int] = None,
        response_time_ms: Optional[float] = None,
        user_id: Optional[UUID] = None,
        organization_id: Optional[UUID] = None,
        action: Optional[str] = None,
        resource_type: Optional[str] = None,
        resource_id: Optional[str] = None,
        metadata: Optional[Dict[str, Any]] = None
    ) -> AuditLog:
        """
        Log an API call
        
        Args:
            db: Database session
            request_method: HTTP method
            request_path: Request path
            request_id: Request ID
            request_ip: Requester IP
            request_user_agent: User agent
            response_status: Response status code
            response_time_ms: Response time in ms
            user_id: User ID
            organization_id: Organization ID
            action: Action performed
            resource_type: Resource type
            resource_id: Resource ID
            metadata: Additional metadata
            
        Returns:
            Created audit log entry
        """
        log_entry = AuditLog(
            request_method=request_method,
            request_path=request_path,
            request_id=request_id,
            request_ip=request_ip,
            request_user_agent=request_user_agent,
            response_status=response_status,
            response_time_ms=response_time_ms,
            user_id=user_id,
            organization_id=organization_id,
            action=action,
            resource_type=resource_type,
            resource_id=resource_id,
            metadata=metadata
        )
        
        db.add(log_entry)
        await db.commit()
        await db.refresh(log_entry)
        return log_entry
    
    async def search_logs(
        self,
        db: AsyncSession,
        *,
        user_id: Optional[UUID] = None,
        organization_id: Optional[UUID] = None,
        action: Optional[str] = None,
        resource_type: Optional[str] = None,
        resource_id: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        skip: int = 0,
        limit: int = 100
    ) -> List[AuditLog]:
        """
        Search audit logs with filtering
        
        Args:
            db: Database session
            user_id: Filter by user ID
            organization_id: Filter by organization ID
            action: Filter by action
            resource_type: Filter by resource type
            resource_id: Filter by resource ID
            start_date: Filter by created_at start
            end_date: Filter by created_at end
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of audit logs
        """
        filters = []
        
        if user_id:
            filters.append(AuditLog.user_id == user_id)
        
        if organization_id:
            filters.append(AuditLog.organization_id == organization_id)
        
        if action:
            filters.append(AuditLog.action == action)
        
        if resource_type:
            filters.append(AuditLog.resource_type == resource_type)
        
        if resource_id:
            filters.append(AuditLog.resource_id == resource_id)
        
        if start_date:
            filters.append(AuditLog.created_at >= start_date)
        
        if end_date:
            filters.append(AuditLog.created_at <= end_date)
        
        query = select(AuditLog)
        
        if filters:
            query = query.where(and_(*filters))
        
        query = query.order_by(desc(AuditLog.created_at)).offset(skip).limit(limit)
        
        result = await db.execute(query)
        return result.scalars().all()