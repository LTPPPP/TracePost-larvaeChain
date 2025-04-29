# tracing.py
from typing import Dict, Any, Optional
from uuid import UUID
from fastapi import APIRouter, Depends, UploadFile, File, Form, Query, HTTPException, BackgroundTasks
from sqlalchemy.ext.asyncio import AsyncSession

from app.api.dependencies import get_db, get_current_user, get_current_active_user
from app.services import tracing as tracing_service
from app.models.user import User
from app.schemas.shipment import DocumentCreate
from app.core.exceptions import ResourceNotFoundError, BlockchainError

router = APIRouter(prefix="/tracing", tags=["Tracing"])

@router.post("/shipments/{shipment_id}/verify", response_model=Dict[str, Any])
async def verify_shipment_on_blockchain(
    shipment_id: UUID,
    network: Optional[str] = Query("ethereum", description="Blockchain network to use"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Register shipment data on blockchain for verification
    """
    try:
        result = await tracing_service.verify_shipment_on_blockchain(
            db=db,
            shipment_id=shipment_id,
            user=current_user,
            network=network
        )
        return result
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BlockchainError as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/events/{event_id}/verify", response_model=Dict[str, Any])
async def verify_event_on_blockchain(
    event_id: UUID,
    network: Optional[str] = Query("ethereum", description="Blockchain network to use"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Register event data on blockchain for verification
    """
    try:
        result = await tracing_service.verify_event_on_blockchain(
            db=db,
            event_id=event_id,
            user=current_user,
            network=network
        )
        return result
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BlockchainError as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/documents/store", response_model=Dict[str, Any])
async def store_document(
    document: UploadFile = File(...),
    filename: str = Form(...),
    content_type: str = Form(...),
    shipment_id: Optional[UUID] = Form(None),
    description: Optional[str] = Form(None),
    metadata: Optional[str] = Form(None),
    network: Optional[str] = Query("ethereum", description="Blockchain network to use"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Store a document hash on blockchain and save document metadata
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
            metadata=metadata or "{}"
        )
        
        result = await tracing_service.store_document_hash(
            db=db,
            document_in=document_in,
            user=current_user,
            document_content=document_content,
            network=network
        )
        
        return {
            "document_id": str(result.id),
            "filename": result.filename,
            "content_type": result.content_type,
            "content_hash": result.content_hash,
            "blockchain_tx_hash": result.blockchain_tx_hash,
            "blockchain_network": result.blockchain_network,
            "blockchain_status": result.blockchain_status,
            "created_at": result.created_at.isoformat()
        }
    except BlockchainError as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/transactions/{tx_hash}/verify", response_model=Dict[str, Any])
async def verify_transaction(
    tx_hash: str,
    network: str = Query(..., description="Blockchain network"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Verify a blockchain transaction and update status
    """
    try:
        result = await tracing_service.verify_blockchain_transaction(
            db=db,
            tx_hash=tx_hash,
            network=network
        )
        return result
    except BlockchainError as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/documents/verify", response_model=Dict[str, Any])
async def verify_document(
    document: UploadFile = File(...),
    network: Optional[str] = Query(None, description="Blockchain network (optional)"),
    db: AsyncSession = Depends(get_db)
):
    """
    Verify a document's authenticity against blockchain records
    """
    try:
        # Read document content
        document_content = await document.read()
        
        result = await tracing_service.verify_document_authenticity(
            db=db,
            document_content=document_content,
            network=network
        )
        return result
    except BlockchainError as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.get("/shipments/{tracking_number}/verify", response_model=Dict[str, Any])
async def verify_shipment_authenticity(
    tracking_number: str,
    db: AsyncSession = Depends(get_db)
):
    """
    Verify a shipment's authenticity and events history against blockchain records
    """
    try:
        result = await tracing_service.verify_shipment_authenticity(
            db=db,
            tracking_number=tracking_number
        )
        return result
    except ResourceNotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BlockchainError as e:
        raise HTTPException(status_code=500, detail=str(e))