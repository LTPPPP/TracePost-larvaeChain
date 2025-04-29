# schemas/alert.py
from pydantic import BaseModel, UUID4, Field, validator
from typing import Optional, List, Dict, Any, Union
from datetime import datetime
from enum import Enum

# Enums
class AlertSeverityEnum(str, Enum):
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"
    CRITICAL = "critical"


class AlertTypeEnum(str, Enum):
    TEMPERATURE = "temperature_breach"
    HUMIDITY = "humidity_breach"
    LOCATION = "location_deviation"
    DELAY = "shipment_delay"
    DAMAGED = "package_damaged"
    SECURITY = "security_breach"
    CUSTOMS = "customs_issue"
    BLOCKCHAIN = "blockchain_verification"
    SYSTEM = "system_alert"


class AlertStatusEnum(str, Enum):
    ACTIVE = "active"
    ACKNOWLEDGED = "acknowledged"
    RESOLVED = "resolved"
    IGNORED = "ignored"


# Alert schemas
class AlertBase(BaseModel):
    """Base schema for alert data"""
    title: str
    description: str
    alert_type: str
    severity: AlertSeverityEnum = AlertSeverityEnum.INFO
    source: str
    source_id: Optional[str] = None
    resource_type: Optional[str] = None
    resource_id: Optional[UUID4] = None
    metadata: Optional[Dict[str, Any]] = None


class AlertCreate(AlertBase):
    """Schema for creating a new alert"""
    status: AlertStatusEnum = AlertStatusEnum.ACTIVE
    user_id: Optional[UUID4] = None


class AlertUpdate(BaseModel):
    """Schema for updating an alert"""
    title: Optional[str] = None
    description: Optional[str] = None
    severity: Optional[AlertSeverityEnum] = None
    status: Optional[AlertStatusEnum] = None
    resolved_at: Optional[datetime] = None
    resolution_notes: Optional[str] = None
    notified: Optional[bool] = None
    notification_sent_at: Optional[datetime] = None
    resolved_by_id: Optional[UUID4] = None
    metadata: Optional[Dict[str, Any]] = None


class AlertInDBBase(AlertBase):
    """Base schema for alert in DB"""
    id: UUID4
    status: AlertStatusEnum
    created_at: datetime
    updated_at: datetime
    resolved_at: Optional[datetime] = None
    resolution_notes: Optional[str] = None
    notified: bool
    notification_sent_at: Optional[datetime] = None
    user_id: Optional[UUID4] = None
    resolved_by_id: Optional[UUID4] = None

    class Config:
        from_attributes = True  # Updated from orm_mode = True for Pydantic v2


class Alert(AlertInDBBase):
    """Schema for alert response"""
    pass


# Alert Rule schemas
class AlertRuleBase(BaseModel):
    """Base schema for alert rule data"""
    name: str
    description: Optional[str] = None
    resource_type: str
    alert_type: str
    condition: Dict[str, Any]
    severity: AlertSeverityEnum = AlertSeverityEnum.INFO
    message_template: str
    auto_resolve: bool = False
    organization_id: Optional[UUID4] = None
    metadata: Optional[Dict[str, Any]] = None


class AlertRuleCreate(AlertRuleBase):
    """Schema for creating a new alert rule"""
    is_active: bool = True


class AlertRuleUpdate(BaseModel):
    """Schema for updating an alert rule"""
    name: Optional[str] = None
    description: Optional[str] = None
    resource_type: Optional[str] = None
    alert_type: Optional[str] = None
    condition: Optional[Dict[str, Any]] = None
    severity: Optional[AlertSeverityEnum] = None
    message_template: Optional[str] = None
    auto_resolve: Optional[bool] = None
    is_active: Optional[bool] = None
    organization_id: Optional[UUID4] = None
    metadata: Optional[Dict[str, Any]] = None


class AlertRuleInDBBase(AlertRuleBase):
    """Base schema for alert rule in DB"""
    id: UUID4
    is_active: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True  # Updated from orm_mode = True for Pydantic v2


class AlertRule(AlertRuleInDBBase):
    """Schema for alert rule response"""
    pass


# Alert Subscription schemas
class AlertSubscriptionBase(BaseModel):
    """Base schema for alert subscription data"""
    alert_type: Optional[str] = None
    resource_type: Optional[str] = None
    resource_id: Optional[UUID4] = None
    min_severity: AlertSeverityEnum = AlertSeverityEnum.INFO
    notification_channels: List[str]
    user_id: Optional[UUID4] = None
    organization_id: Optional[UUID4] = None

    @validator('notification_channels')
    def validate_notification_channels(cls, v):
        if not v:
            raise ValueError("At least one notification channel must be specified")
        valid_channels = ["email", "sms", "webhook", "push", "in_app"]
        for channel in v:
            if channel not in valid_channels:
                raise ValueError(f"Invalid notification channel: {channel}")
        return v

    @validator('user_id', 'organization_id')
    def validate_subscriber(cls, v, values, **kwargs):
        field = kwargs.get('field')
        other_field = 'organization_id' if field == 'user_id' else 'user_id'
        
        # If setting this field and the other field is already set, raise error
        if v is not None and values.get(other_field) is not None:
            raise ValueError("Only one of user_id or organization_id can be set")
        
        # If this is the second field and both are None, raise error
        if v is None and field == 'organization_id' and values.get('user_id') is None:
            raise ValueError("Either user_id or organization_id must be set")
        
        return v


class AlertSubscriptionCreate(AlertSubscriptionBase):
    """Schema for creating a new alert subscription"""
    is_active: bool = True


class AlertSubscriptionUpdate(BaseModel):
    """Schema for updating an alert subscription"""
    alert_type: Optional[str] = None
    resource_type: Optional[str] = None
    resource_id: Optional[UUID4] = None
    min_severity: Optional[AlertSeverityEnum] = None
    notification_channels: Optional[List[str]] = None
    is_active: Optional[bool] = None


class AlertSubscriptionInDBBase(AlertSubscriptionBase):
    """Base schema for alert subscription in DB"""
    id: UUID4
    is_active: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True  # Updated from orm_mode = True for Pydantic v2


class AlertSubscription(AlertSubscriptionInDBBase):
    """Schema for alert subscription response"""
    pass


# Alert summary schemas
class AlertCount(BaseModel):
    """Schema for alert count summary"""
    total: int
    by_severity: Dict[str, int]
    by_status: Dict[str, int]
    by_type: Dict[str, int]


class AlertNotification(BaseModel):
    """Schema for alert notification"""
    alert_id: UUID4
    title: str
    description: str
    severity: AlertSeverityEnum
    created_at: datetime
    resource_type: Optional[str] = None
    resource_id: Optional[UUID4] = None