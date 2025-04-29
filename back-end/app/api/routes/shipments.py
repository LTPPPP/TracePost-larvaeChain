# shipments.py
from typing import Dict, Any, List, Optional
from uuid import UUID
from fastapi import APIRouter, Depends, HTTPException, Query, Path, Body, UploadFile, File, Form
from sqlalchemy.ext.asyncio import AsyncSession

from app.api.dependencies import get_db, get_current_user, get_current_active_user
from app.services import shipment as shipment_service
from app.models.user import User
from app.schemas.shipment import (
    ShipmentCreate, 
    ShipmentUpdate, 
    ShipmentResponse, 
    ShipmentListResponse,
    DocumentResponse,
    DocumentCreate
)
from app.core.exceptions import ResourceNotFoundError, BlockchainError

router = APIRouter(prefix="/shipments", tags=["Shipments"])

@router.post("", response_model=ShipmentResponse)
async def create_shipment(
    shipment_in: ShipmentCreate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Create a new shipment
    """
    try:
        shipment = await shipment_service.create_shipment(
            db=db,
            obj_in=shipment_in,
            user=current_user
        )
        return shipment
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("", response_model=ShipmentListResponse)
async def list_shipments(
    status: Optional[str] = None,
    origin: Optional[str] = None,
    destination: Optional[str] = None,
    tracking_number: Optional[str] = None,
    skip: int = 0,
    limit: int = 100,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    List shipments with optional filters
    """
    try:
        shipments, total = await shipment_service.get_shipments(
            db=db,
            user=current_user,
            status=status,
            origin=origin,
            destination=destination,
            tracking_number=tracking_number,
            skip=skip,
            limit=limit
        )
        return {
            "shipments": shipments,
            "total": total
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/{shipment_id}", response_model=ShipmentResponse)
async def get_shipment(
    shipment_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Get shipment by ID
    """
    try:
        shipment = await shipment_service.get_shipment(
            db=db,
            shipment_id=shipment_id,
            user=current_user
        )
        if not shipment:
            raise HTTPException(status_code=404, detail="Shipment not found")
        return shipment
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.put("/{shipment_id}", response_model=ShipmentResponse)
async def update_shipment(
    shipment_id: UUID,
    shipment_in: ShipmentUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Update a shipment
    """
    try:
        shipment = await shipment_service.update_shipment(
            db=db,
            shipment_id=shipment_id,
            obj_in=shipment_in,
            user=current_user
        )
        return shipment
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.delete("/{shipment_id}", response_model=Dict[str, Any])
async def delete_shipment(
    shipment_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Delete a shipment
    """
    try:
        result = await shipment_service.delete_shipment(
            db=db,
            shipment_id=shipment_id,
            user=current_user
        )
        return {"success": result, "message": "Shipment deleted successfully"}
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/tracking/{tracking_number}", response_model=ShipmentResponse)
async def get_shipment_by_tracking(
    tracking_number: str,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Get shipment by tracking number
    """
    try:
        shipment = await shipment_service.get_shipment_by_tracking(
            db=db,
            tracking_number=tracking_number,
            user=current_user
        )
        if not shipment:
            raise HTTPException(status_code=404, detail="Shipment not found")
        return shipment
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/{shipment_id}/documents", response_model=DocumentResponse)
async def upload_document(
    shipment_id: UUID,
    document: UploadFile = File(...),
    filename: str = Form(...),
    content_type: str = Form(...),
    description: Optional[str] = Form(None),
    metadata: Optional[str] = Form("{}"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Upload a document for a shipment
    """
    try:
        # Read document content
        document_content = await document.read()
        
        # Create document data
        document_in = DocumentCreate(
            filename=filename,
            content_type=content_type,
            shipment_id=shipment_id,
            description=description,
            metadata=metadata
        )
        
        result = await shipment_service.upload_document(
            db=db,
            document_in=document_in,
            document_content=document_content,
            user=current_user
        )
        
        return result
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/{shipment_id}/documents", response_model=List[DocumentResponse])
async def get_shipment_documents(
    shipment_id: UUID,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Get documents associated with a shipment
    """
    try:
        documents = await shipment_service.get_shipment_documents(
            db=db,
            shipment_id=shipment_id,
            user=current_user
        )
        return documents
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))