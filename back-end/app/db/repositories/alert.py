from typing import Optional, List, Dict, Any, Tuple
from uuid import UUID
from datetime import datetime, timedelta

from sqlalchemy import select, func, and_, or_, desc
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.repositories.base import BaseRepository
from app.models.alert import Alert, AlertRule, AlertSubscription
from app.models.user import User
from app.schemas.alert import AlertCreate, AlertUpdate, AlertRuleCreate, AlertRuleUpdate
from app.utils.logger import get_logger

logger = get_logger(__name__)


class AlertRepository(BaseRepository[Alert, AlertCreate, AlertUpdate]):
    """Repository for system-wide Alerts"""
    
    def __init__(self):
        super().__init__(Alert)
    
    async def get_active_alerts(
        self,
        db: AsyncSession,
        *,
        organization_id: Optional[UUID] = None,
        resource_type: Optional[str] = None,
        resource_id: Optional[UUID] = None,
        severity: Optional[str] = None,
        skip: int = 0,
        limit: int = 100
    ) -> List[Alert]:
        """
        Get active alerts with optional filtering
        
        Args:
            db: Database session
            organization_id: Organization ID to filter by
            resource_type: Resource type to filter by
            resource_id: Resource ID to filter by
            severity: Severity level to filter by
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of active alerts
        """
        filters = [Alert.status == "active"]
        
        if organization_id:
            # Filter by alerts created by the organization's users
            filters.append(Alert.user_id.in_(
                select(User.id).where(User.organization_id == organization_id)
            ))
        
        if resource_type:
            filters.append(Alert.resource_type == resource_type)
        
        if resource_id:
            filters.append(Alert.resource_id == resource_id)
        
        if severity:
            filters.append(Alert.severity == severity)
        
        query = select(Alert).where(and_(*filters))
        query = query.order_by(desc(Alert.created_at)).offset(skip).limit(limit)
        
        result = await db.execute(query)
        return result.scalars().all()
    
    async def resolve_alert(
        self,
        db: AsyncSession,
        *,
        alert_id: UUID,
        resolved_by_id: Optional[UUID] = None,
        resolution_notes: Optional[str] = None
    ) -> Optional[Alert]:
        """
        Resolve an alert
        
        Args:
            db: Database session
            alert_id: Alert ID
            resolved_by_id: User ID of the resolver
            resolution_notes: Notes on resolution
            
        Returns:
            Updated alert or None if not found
        """
        alert = await self.get(db, id=alert_id)
        if not alert:
            return None
        
        alert.status = "resolved"
        alert.resolved_at = datetime.utcnow()
        alert.resolved_by_id = resolved_by_id
        
        if resolution_notes:
            alert.resolution_notes = resolution_notes
        
        db.add(alert)
        await db.commit()
        await db.refresh(alert)
        return alert
    
    async def acknowledge_alert(
        self,
        db: AsyncSession,
        *,
        alert_id: UUID,
        user_id: UUID,
    ) -> Optional[Alert]:
        """
        Acknowledge an alert
        
        Args:
            db: Database session
            alert_id: Alert ID
            user_id: User ID of the acknowledger
            
        Returns:
            Updated alert or None if not found
        """
        alert = await self.get(db, id=alert_id)
        if not alert:
            return None
        
        alert.status = "acknowledged"
        
        # Add user_id to metadata to track who acknowledged
        if not alert.metadata:
            alert.metadata = {}
        
        alert.metadata["acknowledged_by"] = str(user_id)
        alert.metadata["acknowledged_at"] = datetime.utcnow().isoformat()
        
        db.add(alert)
        await db.commit()
        await db.refresh(alert)
        return alert
    
    async def get_alert_counts(
        self,
        db: AsyncSession,
        *,
        organization_id: Optional[UUID] = None,
        days: int = 30
    ) -> Dict[str, Any]:
        """
        Get alert counts and statistics
        
        Args:
            db: Database session
            organization_id: Organization ID to filter by
            days: Number of days to include in statistics
            
        Returns:
            Dictionary with alert statistics
        """
        filters = []
        
        if organization_id:
            # Filter by alerts created by the organization's users
            filters.append(Alert.user_id.in_(
                select(User.id).where(User.organization_id == organization_id)
            ))
        
        # For recent alerts, only those created in the last 'days'
        since = datetime.utcnow() - timedelta(days=days)
        recents_filter = filters.copy()
        recents_filter.append(Alert.created_at >= since)
        
        # Get total counts
        total_query = select(func.count(Alert.id))
        if filters:
            total_query = total_query.where(and_(*filters))
        
        total_result = await db.execute(total_query)
        total = total_result.scalar() or 0
        
        # Get counts by severity
        severity_query = select(Alert.severity, func.count(Alert.id))
        if filters:
            severity_query = severity_query.where(and_(*filters))
        
        severity_query = severity_query.group_by(Alert.severity)
        severity_result = await db.execute(severity_query)
        
        severity_counts = {severity: count for severity, count in severity_result.all()}
        
        # Get counts by status
        status_query = select(Alert.status, func.count(Alert.id))
        if filters:
            status_query = status_query.where(and_(*filters))
        
        status_query = status_query.group_by(Alert.status)
        status_result = await db.execute(status_query)
        
        status_counts = {status: count for status, count in status_result.all()}
        
        # Get counts by type
        type_query = select(Alert.alert_type, func.count(Alert.id))
        if filters:
            type_query = type_query.where(and_(*filters))
        
        type_query = type_query.group_by(Alert.alert_type)
        type_result = await db.execute(type_query)
        
        type_counts = {alert_type: count for alert_type, count in type_result.all()}
        
        return {
            "total": total,
            "by_severity": severity_counts,
            "by_status": status_counts,
            "by_type": type_counts
        }


class AlertRuleRepository(BaseRepository[AlertRule, AlertRuleCreate, AlertRuleUpdate]):
    """Repository for AlertRule operations"""
    
    def __init__(self):
        super().__init__(AlertRule)
    
    async def get_active_rules(
        self,
        db: AsyncSession,
        *,
        organization_id: Optional[UUID] = None,
        resource_type: Optional[str] = None,
        alert_type: Optional[str] = None
    ) -> List[AlertRule]:
        """
        Get active alert rules with optional filtering
        
        Args:
            db: Database session
            organization_id: Organization ID to filter by
            resource_type: Resource type to filter by
            alert_type: Alert type to filter by
            
        Returns:
            List of active alert rules
        """
        filters = [AlertRule.is_active == True]
        
        if organization_id:
            # Either rules for this organization or global rules (organization_id is NULL)
            filters.append(
                or_(
                    AlertRule.organization_id == organization_id,
                    AlertRule.organization_id.is_(None)
                )
            )
        
        if resource_type:
            filters.append(AlertRule.resource_type == resource_type)
        
        if alert_type:
            filters.append(AlertRule.alert_type == alert_type)
        
        query = select(AlertRule).where(and_(*filters))
        result = await db.execute(query)
        return result.scalars().all()


class AlertSubscriptionRepository(BaseRepository[AlertSubscription, AlertCreate, AlertUpdate]):
    """Repository for AlertSubscription operations"""
    
    def __init__(self):
        super().__init__(AlertSubscription)
    
    async def get_by_user(
        self,
        db: AsyncSession,
        *,
        user_id: UUID,
        skip: int = 0,
        limit: int = 100
    ) -> List[AlertSubscription]:
        """
        Get alert subscriptions for a user
        
        Args:
            db: Database session
            user_id: User ID
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of alert subscriptions
        """
        query = select(AlertSubscription).where(
            AlertSubscription.user_id == user_id,
            AlertSubscription.is_active == True
        )
        
        query = query.order_by(AlertSubscription.created_at).offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()
    
    async def get_by_organization(
        self,
        db: AsyncSession,
        *,
        organization_id: UUID,
        skip: int = 0,
        limit: int = 100
    ) -> List[AlertSubscription]:
        """
        Get alert subscriptions for an organization
        
        Args:
            db: Database session
            organization_id: Organization ID
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of alert subscriptions
        """
        query = select(AlertSubscription).where(
            AlertSubscription.organization_id == organization_id,
            AlertSubscription.is_active == True
        )
        
        query = query.order_by(AlertSubscription.created_at).offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()
    
    async def get_subscriptions_for_alert(
        self,
        db: AsyncSession,
        *,
        alert_type: str,
        resource_type: Optional[str] = None,
        resource_id: Optional[UUID] = None,
        severity: str = "info"
    ) -> List[AlertSubscription]:
        """
        Find subscriptions that match an alert
        
        Args:
            db: Database session
            alert_type: Alert type
            resource_type: Resource type
            resource_id: Resource ID
            severity: Alert severity
            
        Returns:
            List of matching alert subscriptions
        """
        # Build a complex query to match subscriptions
        # A subscription matches if:
        # 1. It's active
        # 2. The alert_type matches OR subscription.alert_type is NULL (all types)
        # 3. The resource_type matches OR subscription.resource_type is NULL (all resources)
        # 4. The resource_id matches OR subscription.resource_id is NULL (all instances)
        # 5. The severity is >= subscription.min_severity
        
        severity_rank = {
            "info": 0,
            "warning": 1,
            "error": 2,
            "critical": 3
        }
        
        severity_value = severity_rank.get(severity, 0)
        
        # Filter conditions
        conditions = [
            AlertSubscription.is_active == True,
            or_(
                AlertSubscription.alert_type == alert_type,
                AlertSubscription.alert_type.is_(None)
            )
        ]
        
        # Add resource type condition if provided
        if resource_type:
            conditions.append(
                or_(
                    AlertSubscription.resource_type == resource_type,
                    AlertSubscription.resource_type.is_(None)
                )
            )
        
        # Add resource ID condition if provided
        if resource_id:
            conditions.append(
                or_(
                    AlertSubscription.resource_id == resource_id,
                    AlertSubscription.resource_id.is_(None)
                )
            )
        
        # Add severity condition based on ranking
        # We need to manually map the enum values to numeric values for comparison
        severity_condition = or_(
            AlertSubscription.min_severity == "info",
            and_(AlertSubscription.min_severity == "warning", severity_value >= 1),
            and_(AlertSubscription.min_severity == "error", severity_value >= 2),
            and_(AlertSubscription.min_severity == "critical", severity_value >= 3),
        )
        
        conditions.append(severity_condition)
        
        query = select(AlertSubscription).where(and_(*conditions))
        result = await db.execute(query)
        return result.scalars().all()