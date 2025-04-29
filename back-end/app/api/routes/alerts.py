# api/routes/alerts.py
from typing import List, Optional, Tuple, Dict, Any
from uuid import UUID
from datetime import datetime
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.future import select
from sqlalchemy import or_, and_
from sqlalchemy.orm import selectinload
from fastapi import APIRouter, Depends, HTTPException, Query, Path, status
from pydantic import BaseModel

# Import models directly from the app.models.alert module
from app.models.alert import (
    Alert,
    AlertRule,
    AlertSubscription,
    AlertSeverity,
    AlertType,
    AlertStatus
)
from app.models.user import User
from app.schemas.alert import AlertUpdate
from app.core.exceptions import ResourceNotFoundError
from app.api.deps import get_current_user, get_db

# Create router - Make sure to assign to 'router' variable, not 'app'
router = APIRouter()

# Alert service functions
async def get_alerts(
    db: AsyncSession,
    user: User,
    shipment_id: Optional[UUID] = None,
    alert_type: Optional[str] = None,
    status: Optional[str] = None,
    severity: Optional[str] = None,
    skip: int = 0,
    limit: int = 100
) -> Tuple[List[Alert], int]:
    """
    Get alerts with filtering options
    """
    query = select(Alert)
    
    # Apply filters
    if shipment_id:
        query = query.filter(
            and_(
                Alert.resource_type == "shipment",
                Alert.resource_id == shipment_id
            )
        )
    
    if alert_type:
        query = query.filter(Alert.alert_type == alert_type)
    
    if status:
        query = query.filter(Alert.status == status)
    
    if severity:
        query = query.filter(Alert.severity == severity)
    
    # Count total before pagination
    result = await db.execute(query)
    total = len(result.scalars().all())
    
    # Apply pagination
    query = query.offset(skip).limit(limit)
    
    # Execute query
    result = await db.execute(query)
    alerts = result.scalars().all()
    
    return alerts, total

async def get_alert(
    db: AsyncSession,
    alert_id: UUID,
    user: User
) -> Alert:
    """
    Get a specific alert by ID
    """
    result = await db.execute(select(Alert).filter(Alert.id == alert_id))
    alert = result.scalars().first()
    
    if not alert:
        raise ResourceNotFoundError(f"Alert with ID {alert_id} not found")
    
    return alert

async def update_alert(
    db: AsyncSession,
    alert_id: UUID,
    obj_in: AlertUpdate,
    user: User
) -> Alert:
    """
    Update an alert
    """
    alert = await get_alert(db, alert_id, user)
    
    update_data = obj_in.dict(exclude_unset=True)
    
    for field, value in update_data.items():
        setattr(alert, field, value)
    
    # If status is changed to resolved, set resolved_at and resolved_by
    if "status" in update_data and update_data["status"] == AlertStatus.RESOLVED:
        alert.resolved_at = datetime.now()
        alert.resolved_by_id = user.id
    
    db.add(alert)
    await db.commit()
    await db.refresh(alert)
    
    return alert

async def acknowledge_alert(
    db: AsyncSession,
    alert_id: UUID,
    user: User
) -> Alert:
    """
    Acknowledge an alert
    """
    alert = await get_alert(db, alert_id, user)
    
    if alert.status == AlertStatus.ACTIVE:
        alert.status = AlertStatus.ACKNOWLEDGED
        db.add(alert)
        await db.commit()
        await db.refresh(alert)
    
    return alert

async def resolve_alert(
    db: AsyncSession,
    alert_id: UUID,
    resolution_notes: str,
    user: User
) -> Alert:
    """
    Resolve an alert
    """
    alert = await get_alert(db, alert_id, user)
    
    if alert.status in [AlertStatus.ACTIVE, AlertStatus.ACKNOWLEDGED]:
        alert.status = AlertStatus.RESOLVED
        alert.resolution_notes = resolution_notes
        alert.resolved_at = datetime.now()
        alert.resolved_by_id = user.id
        
        db.add(alert)
        await db.commit()
        await db.refresh(alert)
    
    return alert

async def get_shipment_alerts(
    db: AsyncSession,
    shipment_id: UUID,
    user: User,
    status: Optional[str] = None,
    skip: int = 0,
    limit: int = 100
) -> Tuple[List[Alert], int]:
    """
    Get alerts for a specific shipment
    """
    query = select(Alert).filter(
        and_(
            Alert.resource_type == "shipment",
            Alert.resource_id == shipment_id
        )
    )
    
    if status:
        query = query.filter(Alert.status == status)
    
    # Count total before pagination
    result = await db.execute(query)
    total = len(result.scalars().all())
    
    # Apply pagination
    query = query.offset(skip).limit(limit)
    
    # Execute query
    result = await db.execute(query)
    alerts = result.scalars().all()
    
    return alerts, total

# API Endpoints

@router.get("/alerts", response_model=Dict[str, Any])
async def list_alerts(
    shipment_id: Optional[UUID] = None,
    alert_type: Optional[str] = None,
    status: Optional[str] = None,
    severity: Optional[str] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=100),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get a list of alerts with optional filtering
    """
    alerts, total = await get_alerts(
        db=db,
        user=current_user,
        shipment_id=shipment_id,
        alert_type=alert_type,
        status=status,
        severity=severity,
        skip=skip,
        limit=limit
    )
    
    return {
        "items": alerts,
        "total": total
    }

@router.get("/alerts/{alert_id}", response_model=Alert)
async def get_alert_by_id(
    alert_id: UUID = Path(...),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get a specific alert by ID
    """
    try:
        alert = await get_alert(db=db, alert_id=alert_id, user=current_user)
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))

@router.patch("/alerts/{alert_id}", response_model=Alert)
async def update_alert_endpoint(
    update_data: AlertUpdate,
    alert_id: UUID = Path(...),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Update an alert
    """
    try:
        alert = await update_alert(db=db, alert_id=alert_id, obj_in=update_data, user=current_user)
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))

@router.post("/alerts/{alert_id}/acknowledge", response_model=Alert)
async def acknowledge_alert_endpoint(
    alert_id: UUID = Path(...),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Acknowledge an alert
    """
    try:
        alert = await acknowledge_alert(db=db, alert_id=alert_id, user=current_user)
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))

class ResolveAlertRequest(BaseModel):
    resolution_notes: str

@router.post("/alerts/{alert_id}/resolve", response_model=Alert)
async def resolve_alert_endpoint(
    resolve_data: ResolveAlertRequest,
    alert_id: UUID = Path(...),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Resolve an alert
    """
    try:
        alert = await resolve_alert(
            db=db, 
            alert_id=alert_id, 
            resolution_notes=resolve_data.resolution_notes, 
            user=current_user
        )
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(e))

@router.get("/shipments/{shipment_id}/alerts", response_model=Dict[str, Any])
async def list_shipment_alerts(
    shipment_id: UUID = Path(...),
    status: Optional[str] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=100),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get alerts for a specific shipment
    """
    alerts, total = await get_shipment_alerts(
        db=db,
        shipment_id=shipment_id,
        user=current_user,
        status=status,
        skip=skip,
        limit=limit
    )
    
    return {
        "items": alerts,
        "total": total
    }