# alerts.py
from typing import Dict, Any, List, Optional
from uuid import UUID
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.ext.asyncio import AsyncSession

from app.api.dependencies import get_db, get_current_active_user
from app.services import alert as alert_service
from app.models.user import User
from app.schemas.alert import AlertResponse, AlertListResponse, AlertUpdate
from app.core.exceptions import ResourceNotFoundError

router = APIRouter(prefix="/alerts", tags=["Alerts"])

@router.get("", response_model=AlertListResponse)
async def list_alerts(
    shipment_id: Optional[UUID] = None,
    alert_type: Optional[str] = None,
    status: Optional[str] = None,
    severity: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    List alerts with optional filters
    """
    try:
        alerts, total = await alert_service.get_alerts(
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
            "alerts": alerts,
            "total": total
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/{alert_id}", response_model=AlertResponse)
async def get_alert(
    alert_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Get alert by ID
    """
    try:
        alert = await alert_service.get_alert(
            db=db,
            alert_id=alert_id,
            user=current_user
        )
        if not alert:
            raise HTTPException(status_code=404, detail="Alert not found")
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.put("/{alert_id}", response_model=AlertResponse)
async def update_alert(
    alert_id: UUID,
    alert_in: AlertUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Update an alert (status, resolution notes)
    """
    try:
        alert = await alert_service.update_alert(
            db=db,
            alert_id=alert_id,
            obj_in=alert_in,
            user=current_user
        )
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/{alert_id}/acknowledge", response_model=AlertResponse)
async def acknowledge_alert(
    alert_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Acknowledge an alert
    """
    try:
        alert = await alert_service.acknowledge_alert(
            db=db,
            alert_id=alert_id,
            user=current_user
        )
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/{alert_id}/resolve", response_model=AlertResponse)
async def resolve_alert(
    alert_id: UUID,
    resolution_notes: str = Query("", description="Resolution notes"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Mark alert as resolved
    """
    try:
        alert = await alert_service.resolve_alert(
            db=db,
            alert_id=alert_id,
            resolution_notes=resolution_notes,
            user=current_user
        )
        return alert
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/shipment/{shipment_id}", response_model=AlertListResponse)
async def get_shipment_alerts(
    shipment_id: UUID,
    status: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Get alerts for a specific shipment
    """
    try:
        alerts, total = await alert_service.get_shipment_alerts(
            db=db,
            shipment_id=shipment_id,
            user=current_user,
            status=status,
            skip=skip,
            limit=limit
        )
        return {
            "alerts": alerts,
            "total": total
        }
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))