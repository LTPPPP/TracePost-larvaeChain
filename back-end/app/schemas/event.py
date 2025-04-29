from pydantic import BaseModel, UUID4, Field, validator
from typing import Optional, List, Dict, Any, Union
from datetime import datetime
from enum import Enum

# Enums
class EventTypeEnum(str, Enum):
    PICKUP = "pickup"
    DELIVERY = "delivery"
    CUSTOMS_CHECK = "customs_check"
    WAREHOUSE_IN = "warehouse_in"
    WAREHOUSE_OUT = "warehouse_out"
    TRANSIT = "transit"
    DELAY = "delay"
    DOCUMENTATION = "documentation"
    TEMPERATURE_READING = "temperature_reading"
    HUMIDITY_READING = "humidity_reading"
    SHOCK_READING = "shock_reading"
    SECURITY_CHECK = "security_check"
    OTHER = "other"


class EventSourceEnum(str, Enum):
    MANUAL = "manual"
    IOT = "iot"
    GPS = "gps"
    BLOCKCHAIN = "blockchain"
    ORACLE = "oracle"
    SYSTEM = "system"


# Event schemas
class EventBase(BaseModel):
    """Base schema for event data"""
    event_type: str
    location: str
    timestamp: datetime
    description: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    temperature: Optional[float] = None
    humidity: Optional[float] = None
    pressure: Optional[float] = None
    shock: Optional[float] = None
    battery_level: Optional[float] = None
    metadata: Optional[Dict[str, Any]] = None
    source: EventSourceEnum = EventSourceEnum.MANUAL
    verified_by: Optional[str] = None
    signature: Optional[str] = None


class EventCreate(EventBase):
    """Schema for creating a new event"""
    shipment_id: UUID4
    user_id: Optional[UUID4] = None
    organization_id: Optional[UUID4] = None


class EventUpdate(BaseModel):
    """Schema for updating an event"""
    event_type: Optional[str] = None
    location: Optional[str] = None
    timestamp: Optional[datetime] = None
    description: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    temperature: Optional[float] = None
    humidity: Optional[float] = None
    pressure: Optional[float] = None
    shock: Optional[float] = None
    battery_level: Optional[float] = None
    metadata: Optional[Dict[str, Any]] = None
    source: Optional[EventSourceEnum] = None
    verified_by: Optional[str] = None
    signature: Optional[str] = None


class EventInDBBase(EventBase):
    """Base schema for event in DB"""
    id: UUID4
    shipment_id: UUID4
    created_at: datetime
    updated_at: datetime
    user_id: Optional[UUID4] = None
    organization_id: Optional[UUID4] = None
    blockchain_tx_hash: Optional[str] = None
    blockchain_network: Optional[str] = None
    blockchain_timestamp: Optional[datetime] = None
    blockchain_status: str

    class Config:
        from_attributes = True  # Updated from orm_mode to from_attributes for Pydantic v2


class Event(EventInDBBase):
    """Schema for event response"""
    pass


# Add EventResponse schema that was missing
class EventResponse(Event):
    """Schema for event API responses"""
    pass


# Timeline schemas
class TimelineEvent(BaseModel):
    """Schema for timeline event"""
    id: UUID4
    event_type: str
    location: str
    timestamp: datetime
    description: Optional[str] = None
    source: EventSourceEnum
    verified: bool = False
    blockchain_verified: bool = False
    
    class Config:
        from_attributes = True  # Updated from orm_mode to from_attributes for Pydantic v2


class Timeline(BaseModel):
    """Schema for timeline"""
    shipment_id: UUID4
    tracking_number: str
    events: List[TimelineEvent]
    total_events: int
    
    class Config:
        from_attributes = True  # Updated from orm_mode to from_attributes for Pydantic v2


# Add TimelineResponse schema that was missing
class TimelineResponse(BaseModel):
    """Schema for timeline API responses"""
    events: List[TimelineEvent]
    total: int


class EventSummary(BaseModel):
    """Schema for event summary"""
    event_type: str
    count: int
    last_updated: datetime


class ShipmentEventStats(BaseModel):
    """Schema for shipment event statistics"""
    shipment_id: UUID4
    tracking_number: str
    total_events: int
    events_by_type: Dict[str, int]
    events_by_source: Dict[str, int]
    last_event: Optional[TimelineEvent] = None
    blockchain_verified_count: int
    temperature_stats: Optional[Dict[str, float]] = None
    humidity_stats: Optional[Dict[str, float]] = None