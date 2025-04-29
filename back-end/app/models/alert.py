# alert.py
from sqlalchemy import Column, String, DateTime, ForeignKey, JSON, Text, Boolean, Integer, Enum
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID
import uuid
import enum

from app.db.database import Base
from app.models.base import TimestampMixin, UUIDMixin, BaseCRUD


class AlertSeverity(str, enum.Enum):
    INFO = "info"
    WARNING = "warning"
    ERROR = "error"
    CRITICAL = "critical"


class AlertType(str, enum.Enum):
    TEMPERATURE = "temperature_breach"
    HUMIDITY = "humidity_breach"
    LOCATION = "location_deviation"
    DELAY = "shipment_delay"
    DAMAGED = "package_damaged"
    SECURITY = "security_breach"
    CUSTOMS = "customs_issue"
    BLOCKCHAIN = "blockchain_verification"
    SYSTEM = "system_alert"


class AlertStatus(str, enum.Enum):
    ACTIVE = "active"
    ACKNOWLEDGED = "acknowledged"
    RESOLVED = "resolved"
    IGNORED = "ignored"


class Alert(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Alert model for system-wide alerts"""
    __tablename__ = "alerts"

    # Alert details
    title = Column(String, nullable=False)
    description = Column(Text, nullable=False)
    alert_type = Column(String, nullable=False)
    severity = Column(Enum(AlertSeverity), default=AlertSeverity.INFO, nullable=False)
    status = Column(Enum(AlertStatus), default=AlertStatus.ACTIVE, nullable=False)
    
    # Resolution information
    resolved_at = Column(DateTime(timezone=True), nullable=True)
    resolution_notes = Column(Text, nullable=True)
    
    # Alert source
    source = Column(String, nullable=False)  # 'iot', 'gps', 'blockchain', 'system', 'manual', etc.
    source_id = Column(String, nullable=True)  # ID within the source system
    
    # Related resources
    resource_type = Column(String, nullable=True)  # 'shipment', 'event', 'user', etc.
    resource_id = Column(UUID(as_uuid=True), nullable=True)
    
    # Notification status
    notified = Column(Boolean, default=False, nullable=False)
    notification_sent_at = Column(DateTime(timezone=True), nullable=True)
    
    # Additional data
    metadata = Column(JSON, nullable=True)
    
    # Relationships
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=True)
    resolved_by_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=True)
    
    def __repr__(self):
        return f"<Alert {self.id}: {self.title} ({self.status})>"


class AlertRule(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Rules for generating alerts automatically"""
    __tablename__ = "alert_rules"

    name = Column(String, nullable=False)
    description = Column(Text, nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    
    # Rule conditions
    resource_type = Column(String, nullable=False)  # 'shipment', 'event', etc.
    alert_type = Column(String, nullable=False)
    condition = Column(JSON, nullable=False)  # JSON with condition details
    severity = Column(Enum(AlertSeverity), default=AlertSeverity.INFO, nullable=False)
    
    # Action to take
    message_template = Column(Text, nullable=False)
    auto_resolve = Column(Boolean, default=False, nullable=False)
    
    # Organizations this rule applies to (null means all)
    organization_id = Column(UUID(as_uuid=True), ForeignKey("organizations.id"), nullable=True)
    
    # Rule metadata
    metadata = Column(JSON, nullable=True)
    
    def __repr__(self):
        return f"<AlertRule {self.name}>"


class AlertSubscription(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """User or organization subscriptions to alert types"""
    __tablename__ = "alert_subscriptions"

    # What to subscribe to
    alert_type = Column(String, nullable=True)  # Null means all types
    resource_type = Column(String, nullable=True)  # Null means all resource types
    resource_id = Column(UUID(as_uuid=True), nullable=True)  # Specific resource or null for all
    min_severity = Column(Enum(AlertSeverity), default=AlertSeverity.INFO, nullable=False)
    
    # Who is subscribing
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=True)
    organization_id = Column(UUID(as_uuid=True), ForeignKey("organizations.id"), nullable=True)
    
    # Notification settings
    notification_channels = Column(JSON, nullable=False)  # ['email', 'sms', 'webhook', etc.]
    is_active = Column(Boolean, default=True, nullable=False)
    
    # Constraints
    __table_args__ = (
        # Either user_id or organization_id must be set, but not both
        # This would typically be handled with a CheckConstraint, but we'll enforce in application code
    )
    
    def __repr__(self):
        subscriber = f"user:{self.user_id}" if self.user_id else f"org:{self.organization_id}"
        return f"<AlertSubscription {subscriber} for {self.alert_type or 'all types'}>"