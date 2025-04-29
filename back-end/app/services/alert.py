from typing import List, Optional, Dict, Any, Tuple
from uuid import UUID
from datetime import datetime, timedelta
import json

from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends

from app.db.repositories import (
    alert_repository,
    alert_rule_repository,
    alert_subscription_repository,
    shipment_alert_repository
)
from app.schemas.alert import (
    AlertCreate,
    AlertUpdate,
    AlertRuleCreate,
    AlertRuleUpdate,
    AlertSubscriptionCreate,
    AlertSubscriptionUpdate
)
from app.models.alert import Alert, AlertRule, AlertSubscription, AlertSeverity, AlertType, AlertStatus
from app.models.event import ShipmentAlert, ShipmentEvent
from app.models.user import User
from app.core.exceptions import (
    ResourceNotFoundError,
    ValidationError,
    AlertProcessingError
)
from app.utils.logger import get_logger

logger = get_logger(__name__)

async def create_alert(
    db: AsyncSession,
    alert_in: AlertCreate,
    user: Optional[User] = None
) -> Alert:
    """
    Create a new system alert
    
    Args:
        db: Database session
        alert_in: Alert data
        user: Current user (optional)
        
    Returns:
        Created alert
    """
    # Add user ID if provided
    alert_data = alert_in.dict()
    if user:
        alert_data["user_id"] = user.id
    
    # Create alert
    alert = await alert_repository.create(db, obj_in=alert_data)
    logger.info(f"Alert created: {alert.title} (ID: {alert.id})")
    
    # TODO: Trigger notifications based on subscriptions
    
    return alert

async def update_alert(
    db: AsyncSession,
    alert_id: UUID,
    alert_in: AlertUpdate,
    user: User
) -> Alert:
    """
    Update an alert
    
    Args:
        db: Database session
        alert_id: Alert ID
        alert_in: New alert data
        user: Current user
        
    Returns:
        Updated alert
        
    Raises:
        ResourceNotFoundError: If alert not found
    """
    # Get the alert
    alert = await alert_repository.get(db, id=alert_id)
    if not alert:
        logger.warning(f"Alert update failed: Alert {alert_id} not found")
        raise ResourceNotFoundError(detail="Alert not found")
    
    # Update alert
    alert = await alert_repository.update(db, db_obj=alert, obj_in=alert_in)
    logger.info(f"Alert updated: {alert.title} (ID: {alert.id})")
    
    return alert

async def resolve_alert(
    db: AsyncSession,
    alert_id: UUID,
    resolution_notes: Optional[str] = None,
    user: User = None
) -> Alert:
    """
    Resolve an alert
    
    Args:
        db: Database session
        alert_id: Alert ID
        resolution_notes: Notes about resolution
        user: Current user
        
    Returns:
        Resolved alert
        
    Raises:
        ResourceNotFoundError: If alert not found
    """
    # Get the alert
    resolved_by_id = user.id if user else None
    alert = await alert_repository.resolve_alert(
        db, 
        alert_id=alert_id, 
        resolved_by_id=resolved_by_id,
        resolution_notes=resolution_notes
    )
    
    if not alert:
        logger.warning(f"Alert resolution failed: Alert {alert_id} not found")
        raise ResourceNotFoundError(detail="Alert not found")
    
    logger.info(f"Alert resolved: {alert.title} (ID: {alert.id})")
    return alert

async def acknowledge_alert(
    db: AsyncSession,
    alert_id: UUID,
    user: User
) -> Alert:
    """
    Acknowledge an alert
    
    Args:
        db: Database session
        alert_id: Alert ID
        user: Current user
        
    Returns:
        Acknowledged alert
        
    Raises:
        ResourceNotFoundError: If alert not found
    """
    # Get the alert
    alert = await alert_repository.acknowledge_alert(db, alert_id=alert_id, user_id=user.id)
    
    if not alert:
        logger.warning(f"Alert acknowledgement failed: Alert {alert_id} not found")
        raise ResourceNotFoundError(detail="Alert not found")
    
    logger.info(f"Alert acknowledged: {alert.title} (ID: {alert.id})")
    return alert

async def get_alert(
    db: AsyncSession,
    alert_id: UUID
) -> Alert:
    """
    Get an alert by ID
    
    Args:
        db: Database session
        alert_id: Alert ID
        
    Returns:
        Alert if found
        
    Raises:
        ResourceNotFoundError: If alert not found
    """
    alert = await alert_repository.get(db, id=alert_id)
    
    if not alert:
        logger.warning(f"Alert not found: {alert_id}")
        raise ResourceNotFoundError(detail="Alert not found")
    
    return alert

async def get_active_alerts(
    db: AsyncSession,
    user: User,
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
        user: Current user
        resource_type: Filter by resource type
        resource_id: Filter by resource ID
        severity: Filter by alert severity
        skip: Number of records to skip
        limit: Maximum number of records to return
        
    Returns:
        List of alerts
    """
    # If user is admin, get all alerts
    organization_id = None if user.role == "admin" else user.organization_id
    
    return await alert_repository.get_active_alerts(
        db,
        organization_id=organization_id,
        resource_type=resource_type,
        resource_id=resource_id,
        severity=severity,
        skip=skip,
        limit=limit
    )

async def get_alert_counts(
    db: AsyncSession,
    user: User,
    days: int = 30
) -> Dict[str, Any]:
    """
    Get alert counts and statistics
    
    Args:
        db: Database session
        user: Current user
        days: Number of days to include in statistics
        
    Returns:
        Dictionary with alert statistics
    """
    # If user is admin, get all stats
    organization_id = None if user.role == "admin" else user.organization_id
    
    return await alert_repository.get_alert_counts(
        db,
        organization_id=organization_id,
        days=days
    )

async def create_alert_rule(
    db: AsyncSession,
    rule_in: AlertRuleCreate,
    user: User
) -> AlertRule:
    """
    Create a new alert rule
    
    Args:
        db: Database session
        rule_in: Alert rule data
        user: Current user
        
    Returns:
        Created alert rule
    """
    # Add organization ID for non-admin users
    rule_data = rule_in.dict()
    if user.role != "admin":
        rule_data["organization_id"] = user.organization_id
    
    # Create rule
    rule = await alert_rule_repository.create(db, obj_in=rule_data)
    logger.info(f"Alert rule created: {rule.name} (ID: {rule.id})")
    
    return rule

async def update_alert_rule(
    db: AsyncSession,
    rule_id: UUID,
    rule_in: AlertRuleUpdate,
    user: User
) -> AlertRule:
    """
    Update an alert rule
    
    Args:
        db: Database session
        rule_id: Alert rule ID
        rule_in: New alert rule data
        user: Current user
        
    Returns:
        Updated alert rule
        
    Raises:
        ResourceNotFoundError: If alert rule not found
    """
    # Get the rule
    rule = await alert_rule_repository.get(db, id=rule_id)
    if not rule:
        logger.warning(f"Alert rule update failed: Rule {rule_id} not found")
        raise ResourceNotFoundError(detail="Alert rule not found")
    
    # Check if user has access to the rule
    if user.role != "admin" and rule.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to update alert rule from another organization")
        raise ResourceNotFoundError(detail="Alert rule not found")
    
    # Update rule
    rule = await alert_rule_repository.update(db, db_obj=rule, obj_in=rule_in)
    logger.info(f"Alert rule updated: {rule.name} (ID: {rule.id})")
    
    return rule

async def get_alert_rule(
    db: AsyncSession,
    rule_id: UUID,
    user: Optional[User] = None
) -> AlertRule:
    """
    Get an alert rule by ID
    
    Args:
        db: Database session
        rule_id: Alert rule ID
        user: Current user
        
    Returns:
        Alert rule if found
        
    Raises:
        ResourceNotFoundError: If alert rule not found
    """
    rule = await alert_rule_repository.get(db, id=rule_id)
    
    if not rule:
        logger.warning(f"Alert rule not found: {rule_id}")
        raise ResourceNotFoundError(detail="Alert rule not found")
    
    # Check if user has access to the rule
    if user and user.role != "admin" and rule.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to access alert rule from another organization")
        raise ResourceNotFoundError(detail="Alert rule not found")
    
    return rule

async def get_active_alert_rules(
    db: AsyncSession,
    user: User,
    resource_type: Optional[str] = None,
    alert_type: Optional[str] = None
) -> List[AlertRule]:
    """
    Get active alert rules
    
    Args:
        db: Database session
        user: Current user
        resource_type: Filter by resource type
        alert_type: Filter by alert type
        
    Returns:
        List of alert rules
    """
    # If user is admin, get all rules
    organization_id = None if user.role == "admin" else user.organization_id
    
    return await alert_rule_repository.get_active_rules(
        db,
        organization_id=organization_id,
        resource_type=resource_type,
        alert_type=alert_type
    )

async def create_alert_subscription(
    db: AsyncSession,
    subscription_in: AlertSubscriptionCreate,
    user: User
) -> AlertSubscription:
    """
    Create a new alert subscription
    
    Args:
        db: Database session
        subscription_in: Alert subscription data
        user: Current user
        
    Returns:
        Created alert subscription
    """
    # Ensure either user_id or organization_id is set, not both
    subscription_data = subscription_in.dict()
    
    if subscription_data.get("user_id") and subscription_data.get("organization_id"):
        logger.warning("Subscription creation failed: Both user_id and organization_id provided")
        raise ValidationError(detail="Cannot specify both user_id and organization_id")
    
    if not subscription_data.get("user_id") and not subscription_data.get("organization_id"):
        # Default to current user
        subscription_data["user_id"] = user.id
    
    # Create subscription
    subscription = await alert_subscription_repository.create(db, obj_in=subscription_data)
    logger.info(f"Alert subscription created: ID {subscription.id}")
    
    return subscription

async def update_alert_subscription(
    db: AsyncSession,
    subscription_id: UUID,
    subscription_in: AlertSubscriptionUpdate,
    user: User
) -> AlertSubscription:
    """
    Update an alert subscription
    
    Args:
        db: Database session
        subscription_id: Alert subscription ID
        subscription_in: New alert subscription data
        user: Current user
        
    Returns:
        Updated alert subscription
        
    Raises:
        ResourceNotFoundError: If alert subscription not found
    """
    # Get the subscription
    subscription = await alert_subscription_repository.get(db, id=subscription_id)
    if not subscription:
        logger.warning(f"Alert subscription update failed: Subscription {subscription_id} not found")
        raise ResourceNotFoundError(detail="Alert subscription not found")
    
    # Check if user has access to the subscription
    if user.role != "admin" and subscription.user_id != user.id:
        if not (subscription.organization_id and subscription.organization_id == user.organization_id):
            logger.warning(f"User {user.id} attempted to update alert subscription they don't own")
            raise ResourceNotFoundError(detail="Alert subscription not found")
    
    # Update subscription
    subscription = await alert_subscription_repository.update(db, db_obj=subscription, obj_in=subscription_in)
    logger.info(f"Alert subscription updated: ID {subscription.id}")
    
    return subscription

async def get_user_subscriptions(
    db: AsyncSession,
    user: User,
    skip: int = 0,
    limit: int = 100
) -> List[AlertSubscription]:
    """
    Get alert subscriptions for a user
    
    Args:
        db: Database session
        user: Current user
        skip: Number of records to skip
        limit: Maximum number of records to return
        
    Returns:
        List of alert subscriptions
    """
    return await alert_subscription_repository.get_by_user(
        db,
        user_id=user.id,
        skip=skip,
        limit=limit
    )

async def get_organization_subscriptions(
    db: AsyncSession,
    user: User,
    skip: int = 0,
    limit: int = 100
) -> List[AlertSubscription]:
    """
    Get alert subscriptions for an organization
    
    Args:
        db: Database session
        user: Current user
        skip: Number of records to skip
        limit: Maximum number of records to return
        
    Returns:
        List of alert subscriptions
    """
    if not user.organization_id:
        logger.warning(f"User {user.id} has no organization")
        return []
    
    return await alert_subscription_repository.get_by_organization(
        db,
        organization_id=user.organization_id,
        skip=skip,
        limit=limit
    )

async def process_event_alerts(
    db: AsyncSession,
    event: ShipmentEvent
) -> List[ShipmentAlert]:
    """
    Process potential alerts from a shipment event
    
    Args:
        db: Database session
        event: Shipment event
        
    Returns:
        List of created alerts
    """
    created_alerts = []
    
    try:
        # Check temperature breach
        if event.temperature is not None:
            # Define thresholds - in practice, these would be configurable per shipment
            temp_max = 30.0  # Celsius
            temp_min = 0.0   # Celsius
            
            if event.temperature > temp_max:
                # Create temperature breach alert
                alert = ShipmentAlert(
                    shipment_id=event.shipment_id,
                    alert_type="temperature_breach",
                    severity="error",
                    message=f"Temperature too high: {event.temperature}°C (max: {temp_max}°C)",
                    expected_value=str(temp_max),
                    actual_value=str(event.temperature),
                    threshold=f"max={temp_max}",
                    event_id=event.id,
                    resolved=False
                )
                db.add(alert)
                await db.flush()
                created_alerts.append(alert)
                
            elif event.temperature < temp_min:
                # Create temperature breach alert
                alert = ShipmentAlert(
                    shipment_id=event.shipment_id,
                    alert_type="temperature_breach",
                    severity="error",
                    message=f"Temperature too low: {event.temperature}°C (min: {temp_min}°C)",
                    expected_value=str(temp_min),
                    actual_value=str(event.temperature),
                    threshold=f"min={temp_min}",
                    event_id=event.id,
                    resolved=False
                )
                db.add(alert)
                await db.flush()
                created_alerts.append(alert)
        
        # Check humidity breach
        if event.humidity is not None:
            # Define thresholds
            humidity_max = 80.0  # Percent
            humidity_min = 20.0  # Percent
            
            if event.humidity > humidity_max:
                # Create humidity breach alert
                alert = ShipmentAlert(
                    shipment_id=event.shipment_id,
                    alert_type="humidity_breach",
                    severity="warning",
                    message=f"Humidity too high: {event.humidity}% (max: {humidity_max}%)",
                    expected_value=str(humidity_max),
                    actual_value=str(event.humidity),
                    threshold=f"max={humidity_max}",
                    event_id=event.id,
                    resolved=False
                )
                db.add(alert)
                await db.flush()
                created_alerts.append(alert)
                
            elif event.humidity < humidity_min:
                # Create humidity breach alert
                alert = ShipmentAlert(
                    shipment_id=event.shipment_id,
                    alert_type="humidity_breach",
                    severity="warning",
                    message=f"Humidity too low: {event.humidity}% (min: {humidity_min}%)",
                    expected_value=str(humidity_min),
                    actual_value=str(event.humidity),
                    threshold=f"min={humidity_min}",
                    event_id=event.id,
                    resolved=False
                )
                db.add(alert)
                await db.flush()
                created_alerts.append(alert)
        
        # Check shock if available
        if event.shock is not None:
            # Define threshold
            shock_max = 5.0  # G-force
            
            if event.shock > shock_max:
                # Create shock alert
                alert = ShipmentAlert(
                    shipment_id=event.shipment_id,
                    alert_type="package_damaged",
                    severity="critical",
                    message=f"Possible package damage: Shock detected at {event.shock}G (max: {shock_max}G)",
                    expected_value=f"<{shock_max}",
                    actual_value=str(event.shock),
                    threshold=f"max={shock_max}",
                    event_id=event.id,
                    resolved=False
                )
                db.add(alert)
                await db.flush()
                created_alerts.append(alert)
        
        # Commit all alerts
        if created_alerts:
            await db.commit()
            for alert in created_alerts:
                logger.info(f"Created shipment alert: {alert.alert_type} (ID: {alert.id})")
            
            # Here you would normally trigger notifications
        
        return created_alerts
        
    except Exception as e:
        logger.error(f"Error processing event alerts: {str(e)}")
        await db.rollback()
        raise AlertProcessingError(detail=f"Failed to process alerts: {str(e)}")