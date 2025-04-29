from typing import List, Optional, Dict, Any, Tuple, Union
from uuid import UUID
from datetime import datetime
import json
import hashlib

from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends

from app.db.repositories import (
    shipment_repository,
    event_repository,
    document_repository
)
from app.blockchain import (
    ethereum_client,
    substrate_client,
    vietnamchain_client
)
from app.models.shipment import Shipment, Document
from app.models.event import ShipmentEvent
from app.models.user import User
from app.services import shipment as shipment_service
from app.services import event as event_service
from app.schemas.shipment import DocumentCreate
from app.core.exceptions import (
    ResourceNotFoundError,
    ValidationError,
    BlockchainError
)
from app.utils.logger import get_logger

logger = get_logger(__name__)

async def verify_shipment_on_blockchain(
    db: AsyncSession,
    shipment_id: UUID,
    user: User,
    network: str = "ethereum"
) -> Dict[str, Any]:
    """
    Register shipment data on blockchain
    
    Args:
        db: Database session
        shipment_id: Shipment ID
        user: Current user
        network: Blockchain network to use
        
    Returns:
        Dictionary with transaction details
        
    Raises:
        ResourceNotFoundError: If shipment not found
        BlockchainError: If blockchain interaction fails
    """
    # Get the shipment
    shipment = await shipment_repository.get(db, id=shipment_id)
    if not shipment:
        logger.warning(f"Shipment not found: {shipment_id}")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Check user access
    if user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to verify shipment from another organization")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Prepare shipment data
    shipment_data = {
        "shipment_id": str(shipment.id),
        "tracking_number": shipment.tracking_number,
        "organization_id": str(shipment.organization_id),
        "status": shipment.status,
        "created_at": shipment.created_at.isoformat(),
        "origin": shipment.origin,
        "destination": shipment.destination,
        "timestamp": datetime.utcnow().isoformat()
    }
    
    # Calculate hash of shipment data
    data_hash = hashlib.sha256(json.dumps(shipment_data, sort_keys=True).encode()).hexdigest()
    
    # Select blockchain client
    client = None
    if network.lower() == "ethereum":
        client = ethereum_client
    elif network.lower() == "substrate":
        client = substrate_client
    elif network.lower() == "vietnamchain":
        client = vietnamchain_client
    else:
        logger.error(f"Unsupported blockchain network: {network}")
        raise BlockchainError(detail=f"Unsupported blockchain network: {network}")
    
    try:
        # Register on blockchain
        tx_hash = await client.register_shipment(
            shipment_id=str(shipment.id),
            tracking_number=shipment.tracking_number,
            data_hash=data_hash,
            metadata=json.dumps(shipment_data)
        )
        
        # Update shipment with blockchain transaction details
        updated_shipment = await shipment_repository.update_blockchain_status(
            db,
            shipment_id=shipment.id,
            tx_hash=tx_hash,
            network=network,
            status="pending"
        )
        
        logger.info(f"Shipment {shipment.tracking_number} registered on {network} blockchain: {tx_hash}")
        
        return {
            "shipment_id": str(shipment.id),
            "tx_hash": tx_hash,
            "network": network,
            "status": "pending",
            "timestamp": datetime.utcnow().isoformat(),
            "data_hash": data_hash
        }
        
    except Exception as e:
        logger.error(f"Blockchain error: {str(e)}")
        raise BlockchainError(detail=f"Failed to register shipment on blockchain: {str(e)}")

async def verify_event_on_blockchain(
    db: AsyncSession,
    event_id: UUID,
    user: User,
    network: str = "ethereum"
) -> Dict[str, Any]:
    """
    Register event data on blockchain
    
    Args:
        db: Database session
        event_id: Event ID
        user: Current user
        network: Blockchain network to use
        
    Returns:
        Dictionary with transaction details
        
    Raises:
        ResourceNotFoundError: If event not found
        BlockchainError: If blockchain interaction fails
    """
    # Get the event
    event = await event_repository.get(db, id=event_id)
    if not event:
        logger.warning(f"Event not found: {event_id}")
        raise ResourceNotFoundError(detail="Event not found")
    
    # Get associated shipment
    shipment = await shipment_repository.get(db, id=event.shipment_id)
    if not shipment:
        logger.warning(f"Associated shipment not found: {event.shipment_id}")
        raise ResourceNotFoundError(detail="Event not found")
    
    # Check user access
    if user.role != "admin" and shipment.organization_id != user.organization_id:
        logger.warning(f"User {user.id} attempted to verify event from another organization")
        raise ResourceNotFoundError(detail="Event not found")
    
    # Prepare event data
    event_data = {
        "event_id": str(event.id),
        "shipment_id": str(event.shipment_id),
        "tracking_number": shipment.tracking_number,
        "event_type": event.event_type,
        "timestamp": event.timestamp.isoformat(),
        "location": event.location,
        "latitude": event.latitude,
        "longitude": event.longitude,
        "description": event.description,
        "source": event.source,
        "created_at": event.created_at.isoformat(),
        "verification_timestamp": datetime.utcnow().isoformat()
    }
    
    # Add sensor data if available
    if event.temperature is not None:
        event_data["temperature"] = event.temperature
    if event.humidity is not None:
        event_data["humidity"] = event.humidity
    if event.shock is not None:
        event_data["shock"] = event.shock
    
    # Calculate hash of event data
    data_hash = hashlib.sha256(json.dumps(event_data, sort_keys=True).encode()).hexdigest()
    
    # Select blockchain client
    client = None
    if network.lower() == "ethereum":
        client = ethereum_client
    elif network.lower() == "substrate":
        client = substrate_client
    elif network.lower() == "vietnamchain":
        client = vietnamchain_client
    else:
        logger.error(f"Unsupported blockchain network: {network}")
        raise BlockchainError(detail=f"Unsupported blockchain network: {network}")
    
    try:
        # Register on blockchain
        tx_hash = await client.register_event(
            shipment_id=str(event.shipment_id),
            event_id=str(event.id),
            event_type=event.event_type,
            data_hash=data_hash,
            metadata=json.dumps(event_data)
        )
        
        # Update event with blockchain transaction details
        updated_event = await event_repository.update_blockchain_status(
            db,
            event_id=event.id,
            tx_hash=tx_hash,
            network=network,
            status="pending"
        )
        
        logger.info(f"Event {event.id} for shipment {shipment.tracking_number} registered on {network} blockchain: {tx_hash}")
        
        return {
            "event_id": str(event.id),
            "shipment_id": str(event.shipment_id),
            "tx_hash": tx_hash,
            "network": network,
            "status": "pending",
            "timestamp": datetime.utcnow().isoformat(),
            "data_hash": data_hash
        }
        
    except Exception as e:
        logger.error(f"Blockchain error: {str(e)}")
        raise BlockchainError(detail=f"Failed to register event on blockchain: {str(e)}")

async def store_document_hash(
    db: AsyncSession,
    document_in: DocumentCreate,
    user: User,
    document_content: bytes,
    network: str = "ethereum"
) -> Document:
    """
    Store a document hash on blockchain and save document metadata
    
    Args:
        db: Database session
        document_in: Document data
        user: Current user
        document_content: Binary document content
        network: Blockchain network to use
        
    Returns:
        Created document
        
    Raises:
        ValidationError: If document data is invalid
        BlockchainError: If blockchain interaction fails
    """
    # Calculate document hash
    document_hash = hashlib.sha256(document_content).hexdigest()
    
    # Check if document with same hash already exists
    existing_doc = await document_repository.get_by_hash(db, document_hash)
    if existing_doc:
        logger.warning(f"Document with hash {document_hash} already exists")
        return existing_doc
    
    # Prepare document data
    doc_data = {
        "filename": document_in.filename,
        "content_type": document_in.content_type,
        "content_hash": document_hash,
        "size_bytes": len(document_content),
        "metadata": document_in.metadata,
        "timestamp": datetime.utcnow().isoformat()
    }
    
    # Select blockchain client
    client = None
    if network.lower() == "ethereum":
        client = ethereum_client
    elif network.lower() == "substrate":
        client = substrate_client
    elif network.lower() == "vietnamchain":
        client = vietnamchain_client
    else:
        logger.error(f"Unsupported blockchain network: {network}")
        raise BlockchainError(detail=f"Unsupported blockchain network: {network}")
    
    try:
        # Register document hash on blockchain
        tx_hash = await client.register_document(
            document_hash=document_hash,
            metadata=json.dumps(doc_data)
        )
        
        # Create document record
        document_data = document_in.dict()
        document_data["content_hash"] = document_hash
        document_data["size_bytes"] = len(document_content)
        document_data["blockchain_tx_hash"] = tx_hash
        document_data["blockchain_network"] = network
        document_data["blockchain_timestamp"] = datetime.utcnow()
        document_data["blockchain_status"] = "pending"
        document_data["user_id"] = user.id
        
        document = await document_repository.create(db, obj_in=document_data)
        
        logger.info(f"Document {document.filename} ({document_hash}) registered on {network} blockchain: {tx_hash}")
        
        return document
        
    except Exception as e:
        logger.error(f"Blockchain error: {str(e)}")
        raise BlockchainError(detail=f"Failed to register document on blockchain: {str(e)}")

async def verify_blockchain_transaction(
    db: AsyncSession,
    tx_hash: str,
    network: str
) -> Dict[str, Any]:
    """
    Verify a blockchain transaction and update status
    
    Args:
        db: Database session
        tx_hash: Transaction hash
        network: Blockchain network
        
    Returns:
        Transaction details
        
    Raises:
        BlockchainError: If blockchain interaction fails
    """
    # Select blockchain client
    client = None
    if network.lower() == "ethereum":
        client = ethereum_client
    elif network.lower() == "substrate":
        client = substrate_client
    elif network.lower() == "vietnamchain":
        client = vietnamchain_client
    else:
        logger.error(f"Unsupported blockchain network: {network}")
        raise BlockchainError(detail=f"Unsupported blockchain network: {network}")
    
    try:
        # Check transaction status
        tx_status = await client.get_transaction_status(tx_hash)
        
        if tx_status.get("status") == "confirmed":
            # Update corresponding entity status
            entity_type = tx_status.get("entity_type")
            entity_id = tx_status.get("entity_id")
            
            if entity_type == "shipment" and entity_id:
                shipment_id = UUID(entity_id)
                await shipment_service.update_blockchain_status(
                    db,
                    shipment_id=shipment_id,
                    tx_hash=tx_hash,
                    network=network,
                    status="confirmed"
                )
            elif entity_type == "event" and entity_id:
                event_id = UUID(entity_id)
                await event_service.update_blockchain_status(
                    db,
                    event_id=event_id,
                    tx_hash=tx_hash,
                    network=network,
                    status="confirmed"
                )
            # Add other entity types as needed
        
        return tx_status
        
    except Exception as e:
        logger.error(f"Blockchain verification error: {str(e)}")
        raise BlockchainError(detail=f"Failed to verify transaction: {str(e)}")

async def verify_document_authenticity(
    db: AsyncSession,
    document_content: bytes,
    network: Optional[str] = None
) -> Dict[str, Any]:
    """
    Verify a document's authenticity against blockchain records
    
    Args:
        db: Database session
        document_content: Binary document content
        network: Blockchain network (optional)
        
    Returns:
        Verification result
        
    Raises:
        BlockchainError: If blockchain interaction fails
    """
    # Calculate document hash
    document_hash = hashlib.sha256(document_content).hexdigest()
    
    # Find document by hash
    document = await document_repository.get_by_hash(db, document_hash)
    if not document:
        return {
            "verified": False,
            "hash": document_hash,
            "reason": "Document not found in database"
        }
    
    # If network specified, use that network
    if network:
        networks = [network]
    else:
        # Otherwise check all networks where document was registered
        networks = []
        if document.blockchain_network:
            networks.append(document.blockchain_network)
        else:
            networks = ["ethereum", "substrate", "vietnamchain"]
    
    verified = False
    verification_details = {}
    
    for net in networks:
        # Select blockchain client
        client = None
        if net.lower() == "ethereum":
            client = ethereum_client
        elif net.lower() == "substrate":
            client = substrate_client
        elif net.lower() == "vietnamchain":
            client = vietnamchain_client
        else:
            continue
        
        try:
            # Check if document hash exists on blockchain
            verification = await client.verify_document(
                document_hash=document_hash
            )
            
            if verification.get("verified"):
                verified = True
                verification_details[net] = verification
                break
                
        except Exception as e:
            logger.error(f"Document verification error on {net}: {str(e)}")
            verification_details[net] = {"error": str(e)}
    
    return {
        "verified": verified,
        "hash": document_hash,
        "document_id": str(document.id) if document else None,
        "filename": document.filename if document else None,
        "content_type": document.content_type if document else None,
        "created_at": document.created_at.isoformat() if document else None,
        "blockchain_tx_hash": document.blockchain_tx_hash if document else None,
        "blockchain_network": document.blockchain_network if document else None,
        "blockchain_status": document.blockchain_status if document else None,
        "verification_details": verification_details
    }

async def verify_shipment_authenticity(
    db: AsyncSession,
    tracking_number: str
) -> Dict[str, Any]:
    """
    Verify a shipment's authenticity and events history against blockchain records
    
    Args:
        db: Database session
        tracking_number: Shipment tracking number
        
    Returns:
        Verification result with events history
        
    Raises:
        ResourceNotFoundError: If shipment not found
        BlockchainError: If blockchain interaction fails
    """
    # Get the shipment
    shipment = await shipment_repository.get_by_tracking_number(db, tracking_number)
    if not shipment:
        logger.warning(f"Shipment not found with tracking number: {tracking_number}")
        raise ResourceNotFoundError(detail="Shipment not found")
    
    # Get events
    events, _ = await event_repository.get_timeline(db, shipment_id=shipment.id)
    
    # Verify blockchain status
    blockchain_verified = False
    verification_details = {}
    
    if shipment.blockchain_tx_hash and shipment.blockchain_network:
        try:
            # Check shipment blockchain status
            client = None
            if shipment.blockchain_network.lower() == "ethereum":
                client = ethereum_client
            elif shipment.blockchain_network.lower() == "substrate":
                client = substrate_client
            elif shipment.blockchain_network.lower() == "vietnamchain":
                client = vietnamchain_client
                
            if client:
                verification = await client.verify_shipment(
                    shipment_id=str(shipment.id),
                    tracking_number=shipment.tracking_number
                )
                
                blockchain_verified = verification.get("verified", False)
                verification_details["shipment"] = verification
        except Exception as e:
            logger.error(f"Blockchain verification error: {str(e)}")
            verification_details["shipment_error"] = str(e)
    
    # Verify events
    event_verifications = []
    
    for event in events:
        if event.blockchain_tx_hash and event.blockchain_network:
            try:
                # Select client for event
                client = None
                if event.blockchain_network.lower() == "ethereum":
                    client = ethereum_client
                elif event.blockchain_network.lower() == "substrate":
                    client = substrate_client
                elif event.blockchain_network.lower() == "vietnamchain":
                    client = vietnamchain_client
                
                if client:
                    event_verification = await client.verify_event(
                        shipment_id=str(shipment.id),
                        event_id=str(event.id)
                    )
                    
                    event_verifications.append({
                        "event_id": str(event.id),
                        "event_type": event.event_type,
                        "timestamp": event.timestamp.isoformat(),
                        "blockchain_network": event.blockchain_network,
                        "blockchain_tx_hash": event.blockchain_tx_hash,
                        "verified": event_verification.get("verified", False),
                        "details": event_verification
                    })
            except Exception as e:
                logger.error(f"Event blockchain verification error: {str(e)}")
                event_verifications.append({
                    "event_id": str(event.id),
                    "event_type": event.event_type,
                    "timestamp": event.timestamp.isoformat(),
                    "blockchain_network": event.blockchain_network,
                    "blockchain_tx_hash": event.blockchain_tx_hash,
                    "verified": False,
                    "error": str(e)
                })
    
    # Construct response
    timeline_events = []
    verified_events_count = 0
    
    for event in events:
        verified = any(
            v.get("event_id") == str(event.id) and v.get("verified", False)
            for v in event_verifications
        )
        
        if verified:
            verified_events_count += 1
        
        timeline_events.append({
            "id": str(event.id),
            "event_type": event.event_type,
            "timestamp": event.timestamp.isoformat(),
            "location": event.location,
            "description": event.description,
            "source": event.source,
            "blockchain_verified": verified,
            "user_verified": bool(event.verified_by)
        })
    
    return {
        "tracking_number": tracking_number,
        "shipment_id": str(shipment.id),
        "status": shipment.status,
        "origin": shipment.origin,
        "destination": shipment.destination,
        "created_at": shipment.created_at.isoformat(),
        "blockchain_verified": blockchain_verified,
        "blockchain_network": shipment.blockchain_network,
        "blockchain_tx_hash": shipment.blockchain_tx_hash,
        "events": timeline_events,
        "total_events": len(events),
        "verified_events": verified_events_count,
        "verification_details": verification_details,
        "event_verifications": event_verifications
    }