from pydantic import BaseModel, UUID4, Field, validator
from typing import Optional, List, Dict, Any, Union
from datetime import datetime
from enum import Enum

# Enums
class ShipmentStatusEnum(str, Enum):
    CREATED = "created"
    PICKUP_SCHEDULED = "pickup_scheduled"
    PICKED_UP = "picked_up"
    IN_TRANSIT = "in_transit"
    CUSTOMS = "customs"
    DELIVERED = "delivered"
    CANCELLED = "cancelled"
    DELAYED = "delayed"
    RETURNED = "returned"


class VerificationStatusEnum(str, Enum):
    UNVERIFIED = "unverified"
    VERIFIED = "verified"
    DISPUTED = "disputed"


# Shipment schemas
class ShipmentBase(BaseModel):
    """Base schema for shipment data"""
    title: str
    tracking_number: str
    description: Optional[str] = None
    origin: str
    destination: str
    estimated_delivery: Optional[datetime] = None
    weight: Optional[float] = None
    volume: Optional[float] = None
    package_count: int = 1
    is_international: bool = False
    metadata: Optional[Dict[str, Any]] = None


class ShipmentCreate(ShipmentBase):
    """Schema for creating a new shipment"""
    organization_id: UUID4
    status: ShipmentStatusEnum = ShipmentStatusEnum.CREATED
    verification_status: VerificationStatusEnum = VerificationStatusEnum.UNVERIFIED
    barcode: Optional[str] = None
    qr_code: Optional[str] = None
    customs_cleared: bool = False


class ShipmentUpdate(BaseModel):
    """Schema for updating a shipment"""
    title: Optional[str] = None
    description: Optional[str] = None
    status: Optional[ShipmentStatusEnum] = None
    origin: Optional[str] = None
    destination: Optional[str] = None
    estimated_delivery: Optional[datetime] = None
    actual_delivery: Optional[datetime] = None
    weight: Optional[float] = None
    volume: Optional[float] = None
    package_count: Optional[int] = None
    customs_cleared: Optional[bool] = None
    is_international: Optional[bool] = None
    barcode: Optional[str] = None
    qr_code: Optional[str] = None
    verification_status: Optional[VerificationStatusEnum] = None
    metadata: Optional[Dict[str, Any]] = None


class ShipmentInDBBase(ShipmentBase):
    """Base schema for shipment in DB"""
    id: UUID4
    status: ShipmentStatusEnum
    organization_id: UUID4
    created_at: datetime
    updated_at: datetime
    actual_delivery: Optional[datetime] = None
    customs_cleared: bool
    verification_status: VerificationStatusEnum
    barcode: Optional[str] = None
    qr_code: Optional[str] = None
    blockchain_tx_hash: Optional[str] = None
    blockchain_network: Optional[str] = None
    blockchain_timestamp: Optional[datetime] = None
    blockchain_status: str

    class Config:
        orm_mode = True


class Shipment(ShipmentInDBBase):
    """Schema for shipment response"""
    pass


# Shipment Item schemas
class ShipmentItemBase(BaseModel):
    """Base schema for shipment item data"""
    name: str
    description: Optional[str] = None
    quantity: int = 1
    unit_price: Optional[float] = None
    weight: Optional[float] = None
    sku: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None


class ShipmentItemCreate(ShipmentItemBase):
    """Schema for creating a new shipment item"""
    shipment_id: UUID4


class ShipmentItemUpdate(BaseModel):
    """Schema for updating a shipment item"""
    name: Optional[str] = None
    description: Optional[str] = None
    quantity: Optional[int] = None
    unit_price: Optional[float] = None
    weight: Optional[float] = None
    sku: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None


class ShipmentItemInDBBase(ShipmentItemBase):
    """Base schema for shipment item in DB"""
    id: UUID4
    shipment_id: UUID4
    created_at: datetime
    updated_at: datetime

    class Config:
        orm_mode = True


class ShipmentItem(ShipmentItemInDBBase):
    """Schema for shipment item response"""
    pass


# Document schemas
class DocumentBase(BaseModel):
    """Base schema for document data"""
    name: str
    document_type: str
    file_path: Optional[str] = None
    content_hash: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None


class DocumentCreate(DocumentBase):
    """Schema for creating a new document"""
    pass


class DocumentUpdate(BaseModel):
    """Schema for updating a document"""
    name: Optional[str] = None
    document_type: Optional[str] = None
    file_path: Optional[str] = None
    content_hash: Optional[str] = None
    metadata: Optional[Dict[str, Any]] = None


class DocumentInDBBase(DocumentBase):
    """Base schema for document in DB"""
    id: UUID4
    created_at: datetime
    updated_at: datetime
    blockchain_tx_hash: Optional[str] = None
    blockchain_network: Optional[str] = None
    blockchain_timestamp: Optional[datetime] = None
    blockchain_status: str

    class Config:
        orm_mode = True


class Document(DocumentInDBBase):
    """Schema for document response"""
    pass


# Tracking response schemas
class ShipmentWithItems(Shipment):
    """Schema for shipment with items"""
    items: List[ShipmentItem] = []


class ShipmentWithEvents(Shipment):
    """Schema for shipment with events"""
    events: List["Event"] = []  # Forward reference to Event schema


class ShipmentWithItemsAndEvents(Shipment):
    """Schema for shipment with items and events"""
    items: List[ShipmentItem] = []
    events: List["Event"] = []  # Forward reference to Event schema


class ShipmentSummary(BaseModel):
    """Schema for shipment summary"""
    id: UUID4
    title: str
    tracking_number: str
    status: ShipmentStatusEnum
    origin: str
    destination: str
    estimated_delivery: Optional[datetime] = None
    created_at: datetime
    verification_status: VerificationStatusEnum
    
    class Config:
        orm_mode = True