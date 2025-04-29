from typing import Optional, List, Dict, Any
from uuid import UUID
from datetime import datetime

from sqlalchemy import select, func, and_, or_, desc
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import joinedload

from app.db.repositories.base import BaseRepository
from app.models.shipment import Shipment, ShipmentItem, Document
from app.schemas.shipment import ShipmentCreate, ShipmentUpdate, ShipmentItemCreate, ShipmentItemUpdate
from app.utils.logger import get_logger

logger = get_logger(__name__)


class ShipmentRepository(BaseRepository[Shipment, ShipmentCreate, ShipmentUpdate]):
    """Repository for Shipment operations"""
    
    def __init__(self):
        super().__init__(Shipment)
    
    async def get_with_details(self, db: AsyncSession, id: UUID) -> Optional[Shipment]:
        """
        Get a shipment with items and events
        
        Args:
            db: Database session
            id: Shipment ID
            
        Returns:
            Shipment with items and events if found, None otherwise
        """
        query = select(Shipment).where(Shipment.id == id).options(
            joinedload(Shipment.items),
            joinedload(Shipment.events),
            joinedload(Shipment.documents)
        )
        result = await db.execute(query)
        return result.scalars().first()
    
    async def get_by_tracking_number(self, db: AsyncSession, tracking_number: str) -> Optional[Shipment]:
        """
        Get a shipment by tracking number
        
        Args:
            db: Database session
            tracking_number: Shipment tracking number
            
        Returns:
            Shipment if found, None otherwise
        """
        query = select(Shipment).where(Shipment.tracking_number == tracking_number)
        result = await db.execute(query)
        return result.scalars().first()
    
    async def get_by_organization(
        self,
        db: AsyncSession,
        *,
        organization_id: UUID,
        skip: int = 0,
        limit: int = 100,
        status: Optional[str] = None
    ) -> List[Shipment]:
        """
        Get shipments by organization
        
        Args:
            db: Database session
            organization_id: Organization ID
            skip: Number of records to skip
            limit: Maximum number of records to return
            status: Filter by status
            
        Returns:
            List of shipments
        """
        query = select(Shipment).where(Shipment.organization_id == organization_id)
        
        if status:
            query = query.where(Shipment.status == status)
        
        query = query.order_by(desc(Shipment.created_at)).offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()
    
    async def search(
        self,
        db: AsyncSession,
        *,
        organization_id: Optional[UUID] = None,
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
            organization_id: Organization ID
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
        filters = []
        
        if organization_id:
            filters.append(Shipment.organization_id == organization_id)
        
        if query:
            filters.append(
                or_(
                    Shipment.title.ilike(f'%{query}%'),
                    Shipment.tracking_number.ilike(f'%{query}%'),
                    Shipment.description.ilike(f'%{query}%')
                )
            )
        
        if status:
            filters.append(Shipment.status == status)
        
        if start_date:
            filters.append(Shipment.created_at >= start_date)
        
        if end_date:
            filters.append(Shipment.created_at <= end_date)
        
        if is_international is not None:
            filters.append(Shipment.is_international == is_international)
        
        query = select(Shipment)
        
        if filters:
            query = query.where(and_(*filters))
        
        query = query.order_by(desc(Shipment.created_at)).offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()
    
    async def get_shipment_counts(
        self,
        db: AsyncSession,
        *,
        organization_id: Optional[UUID] = None
    ) -> Dict[str, int]:
        """
        Get counts of shipments by status
        
        Args:
            db: Database session
            organization_id: Organization ID to filter by
            
        Returns:
            Dictionary of status -> count
        """
        filters = []
        if organization_id:
            filters.append(Shipment.organization_id == organization_id)
        
        query = select(
            Shipment.status,
            func.count(Shipment.id)
        ).group_by(Shipment.status)
        
        if filters:
            query = query.where(and_(*filters))
        
        result = await db.execute(query)
        counts = {status: count for status, count in result.all()}
        
        # Ensure all statuses are present
        all_statuses = [
            "created", "pickup_scheduled", "picked_up", "in_transit", 
            "customs", "delivered", "cancelled", "delayed", "returned"
        ]
        
        return {status: counts.get(status, 0) for status in all_statuses}
    
    async def update_blockchain_status(
        self,
        db: AsyncSession,
        *,
        shipment_id: UUID,
        tx_hash: str,
        network: str,
        status: str = "pending"
    ) -> Optional[Shipment]:
        """
        Update blockchain status for a shipment
        
        Args:
            db: Database session
            shipment_id: Shipment ID
            tx_hash: Blockchain transaction hash
            network: Blockchain network
            status: Transaction status
            
        Returns:
            Updated shipment or None if not found
        """
        shipment = await self.get(db, id=shipment_id)
        if not shipment:
            return None
        
        shipment.blockchain_tx_hash = tx_hash
        shipment.blockchain_network = network
        shipment.blockchain_timestamp = datetime.utcnow()
        shipment.blockchain_status = status
        
        db.add(shipment)
        await db.commit()
        await db.refresh(shipment)
        return shipment


class ShipmentItemRepository(BaseRepository):
    """Repository for ShipmentItem operations"""
    
    def __init__(self):
        super().__init__(ShipmentItem)
    
    async def get_by_shipment(
        self,
        db: AsyncSession,
        *,
        shipment_id: UUID,
    ) -> List[ShipmentItem]:
        """
        Get items for a shipment
        
        Args:
            db: Database session
            shipment_id: Shipment ID
            
        Returns:
            List of items
        """
        query = select(ShipmentItem).where(ShipmentItem.shipment_id == shipment_id)
        result = await db.execute(query)
        return result.scalars().all()


class DocumentRepository(BaseRepository):
    """Repository for Document operations"""
    
    def __init__(self):
        super().__init__(Document)
    
    async def get_by_shipment(
        self,
        db: AsyncSession,
        *,
        shipment_id: UUID,
    ) -> List[Document]:
        """
        Get documents for a shipment
        
        Args:
            db: Database session
            shipment_id: Shipment ID
            
        Returns:
            List of documents
        """
        query = select(Document).join(
            Document.shipments
        ).where(Shipment.id == shipment_id)
        
        result = await db.execute(query)
        return result.scalars().all()
    
    async def get_by_hash(self, db: AsyncSession, content_hash: str) -> Optional[Document]:
        """
        Get a document by content hash
        
        Args:
            db: Database session
            content_hash: Document content hash
            
        Returns:
            Document if found, None otherwise
        """
        query = select(Document).where(Document.content_hash == content_hash)
        result = await db.execute(query)
        return result.scalars().first()