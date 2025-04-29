# events.py
from typing import Dict, Any, List, Optional
from uuid import UUID
from datetime import datetime, timezone
from fastapi import APIRouter, Depends, HTTPException, Query, Path, Body
from sqlalchemy.ext.asyncio import AsyncSession

from app.api.dependencies import get_db, get_current_user, get_current_active_user
from app.services import event as event_service
from app.models.user import User
from app.schemas.event import EventCreate, EventUpdate, EventResponse, TimelineResponse
from app.core.exceptions import ResourceNotFoundError, BlockchainError

router = APIRouter(prefix="/events", tags=["Events"])

@router.post("", response_model=EventResponse)
async def create_event(
    event_in: EventCreate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Create a new shipment event
    """
    try:
        event = await event_service.create_event(
            db=db,
            obj_in=event_in,
            user=current_user
        )
        return event
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/shipment/{shipment_id}", response_model=TimelineResponse)
async def get_shipment_timeline(
    shipment_id: UUID,
    skip: int = 0,
    limit: int = 100,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Get timeline of events for a shipment
    """
    try:
        events, total = await event_service.get_shipment_timeline(
            db=db,
            shipment_id=shipment_id,
            user=current_user,
            skip=skip,
            limit=limit
        )
        return {
            "events": events,
            "total": total
        }
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/{event_id}", response_model=EventResponse)
async def get_event(
    event_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Get an event by ID
    """
    try:
        event = await event_service.get_event(
            db=db,
            event_id=event_id,
            user=current_user
        )
        if not event:
            raise HTTPException(status_code=404, detail="Event not found")
        return event
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.put("/{event_id}", response_model=EventResponse)
async def update_event(
    event_id: UUID,
    event_in: EventUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Update an event
    """
    try:
        event = await event_service.update_event(
            db=db,
            event_id=event_id,
            obj_in=event_in,
            user=current_user
        )
        return event
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.delete("/{event_id}", response_model=Dict[str, Any])
async def delete_event(
    event_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Delete an event
    """
    try:
        result = await event_service.delete_event(
            db=db,
            event_id=event_id,
            user=current_user
        )
        return {"success": result, "message": "Event deleted successfully"}
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/{event_id}/verify", response_model=EventResponse)
async def verify_event(
    event_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Mark an event as verified by the current user
    """
    try:
        event = await event_service.verify_event(
            db=db,
            event_id=event_id,
            user=current_user
        )
        return event
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))