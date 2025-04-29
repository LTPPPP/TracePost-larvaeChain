# event.py
from sqlalchemy import Column, String, Float, DateTime, ForeignKey, JSON, Text, Boolean
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID
import uuid

from app.db.database import Base
from app.models.base import TimestampMixin, UUIDMixin, BlockchainMixin, BaseCRUD

class ShipmentEvent(Base, UUIDMixin, TimestampMixin, BlockchainMixin, BaseCRUD):
    """Events related to shipments"""
    __tablename__ = "shipment_events"

    # Event details
    event_type = Column(String, nullable=False)  # pickup, delivery, customs_check, warehouse_in, warehouse_out, etc.
    location = Column(String, nullable=False)
    timestamp = Column(DateTime(timezone=True), nullable=False)
    description = Column(Text, nullable=True)
    
    # GPS coordinates
    latitude = Column(Float, nullable=True)
    longitude = Column(Float, nullable=True)
    
    # IoT data
    temperature = Column(Float, nullable=True)
    humidity = Column(Float, nullable=True)
    pressure = Column(Float, nullable=True)
    shock = Column(Float, nullable=True)  # G-force
    battery_level = Column(Float, nullable=True)  # percentage
    
    # Additional data
    metadata = Column(JSON, nullable=True)
    verified_by = Column(String, nullable=True)  # user ID or system name that verified this event
    signature = Column(String, nullable=True)  # digital signature if applicable
    
    # Source of the event
    source = Column(String, default="manual", nullable=False)  # manual, iot, gps, blockchain, oracle, etc.
    
    # Relationships
    shipment_id = Column(UUID(as_uuid=True), ForeignKey("shipments.id"), nullable=False)
    shipment = relationship("Shipment", back_populates="events")
    
    # User who created the event (optional)
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=True)
    
    # Organization who created/registered the event
    organization_id = Column(UUID(as_uuid=True), ForeignKey("organizations.id"), nullable=True)
    
    def __repr__(self):
        return f"<ShipmentEvent {self.event_type} at {self.timestamp}>"

class ShipmentAlert(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Alerts generated for shipments"""
    __tablename__ = "shipment_alerts"

    # Alert details
    alert_type = Column(String, nullable=False)  # delay, temperature_breach, location_deviation, etc.
    severity = Column(String, default="info", nullable=False)  # info, warning, error, critical
    message = Column(Text, nullable=False)
    resolved = Column(Boolean, default=False, nullable=False)
    resolved_at = Column(DateTime(timezone=True), nullable=True)
    
    # Alert data
    expected_value = Column(String, nullable=True)
    actual_value = Column(String, nullable=True)
    threshold = Column(String, nullable=True)
    metadata = Column(JSON, nullable=True)
    
    # Related event that triggered the alert (if applicable)
    event_id = Column(UUID(as_uuid=True), ForeignKey("shipment_events.id"), nullable=True)
    
    # Relationships
    shipment_id = Column(UUID(as_uuid=True), ForeignKey("shipments.id"), nullable=False)
    shipment = relationship("Shipment", back_populates="alerts")
    
    # User who resolved the alert (if applicable)
    resolved_by = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=True)
    
    def __repr__(self):
        return f"<ShipmentAlert {self.alert_type}: {self.message}>"

class AuditLog(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Audit log for API calls and system events"""
    __tablename__ = "audit_logs"

    # Request details
    request_method = Column(String, nullable=False)  # GET, POST, PUT, DELETE, etc.
    request_path = Column(String, nullable=False)
    request_id = Column(String, index=True, nullable=True)
    request_ip = Column(String, nullable=True)
    request_user_agent = Column(String, nullable=True)
    
    # Response details
    response_status = Column(Integer, nullable=True)
    response_time_ms = Column(Float, nullable=True)
    
    # User information
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=True)
    organization_id = Column(UUID(as_uuid=True), ForeignKey("organizations.id"), nullable=True)
    
    # Additional context
    action = Column(String, nullable=True)  # create_shipment, log_event, etc.
    resource_type = Column(String, nullable=True)  # shipment, event, user, etc.
    resource_id = Column(String, nullable=True)
    metadata = Column(JSON, nullable=True)
    
    def __repr__(self):
        return f"<AuditLog {self.request_method} {self.request_path}>"