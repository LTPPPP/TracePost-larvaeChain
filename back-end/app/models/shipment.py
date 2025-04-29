# shipment.py
from sqlalchemy import Column, String, Integer, Float, Boolean, DateTime, ForeignKey, JSON, Table, Text
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID
import uuid

from app.db.database import Base
from app.models.base import TimestampMixin, UUIDMixin, BlockchainMixin, BaseCRUD

# Association table for shipment to documents
shipment_documents = Table(
    "shipment_documents",
    Base.metadata,
    Column("shipment_id", UUID(as_uuid=True), ForeignKey("shipments.id"), primary_key=True),
    Column("document_id", UUID(as_uuid=True), ForeignKey("documents.id"), primary_key=True),
)

class Shipment(Base, UUIDMixin, TimestampMixin, BlockchainMixin, BaseCRUD):
    """Shipment model for tracking goods"""
    __tablename__ = "shipments"

    # Basic information
    title = Column(String, nullable=False)
    tracking_number = Column(String, unique=True, index=True, nullable=False)
    description = Column(Text, nullable=True)
    status = Column(String, default="created", nullable=False)  # created, in_transit, delivered, customs, etc.
    
    # Shipment details
    origin = Column(String, nullable=False)
    destination = Column(String, nullable=False)
    estimated_delivery = Column(DateTime(timezone=True), nullable=True)
    actual_delivery = Column(DateTime(timezone=True), nullable=True)
    
    # Physical characteristics
    weight = Column(Float, nullable=True)  # in kg
    volume = Column(Float, nullable=True)  # in cubic meters
    package_count = Column(Integer, default=1, nullable=False)
    
    # Additional metadata
    shipment_data = Column(JSON, nullable=True)  # Changed from 'metadata' to 'shipment_data'
    customs_cleared = Column(Boolean, default=False, nullable=False)
    is_international = Column(Boolean, default=False, nullable=False)
    
    # QR/Barcode identifiers 
    barcode = Column(String, nullable=True)
    qr_code = Column(String, nullable=True)
    
    # Blockchain verification
    verification_status = Column(String, default="unverified", nullable=False)  # unverified, verified, disputed
    
    # Relationships
    organization_id = Column(UUID(as_uuid=True), ForeignKey("organizations.id"), nullable=False)
    organization = relationship("Organization", back_populates="shipments")
    
    # Shipment items
    items = relationship("ShipmentItem", back_populates="shipment", cascade="all, delete-orphan")
    
    # Shipment events
    events = relationship("ShipmentEvent", back_populates="shipment", cascade="all, delete-orphan")
    
    # Shipment alerts
    alerts = relationship("ShipmentAlert", back_populates="shipment", cascade="all, delete-orphan")
    
    # Shipment documents
    documents = relationship("Document", secondary=shipment_documents, back_populates="shipments")
    
    def __repr__(self):
        return f"<Shipment {self.tracking_number}>"

class ShipmentItem(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Individual items within a shipment"""
    __tablename__ = "shipment_items"

    name = Column(String, nullable=False)
    description = Column(Text, nullable=True)
    quantity = Column(Integer, default=1, nullable=False)
    unit_price = Column(Float, nullable=True)
    weight = Column(Float, nullable=True)  # in kg
    sku = Column(String, nullable=True)
    item_data = Column(JSON, nullable=True)  # Changed from 'metadata' to 'item_data'
    
    # Relationships
    shipment_id = Column(UUID(as_uuid=True), ForeignKey("shipments.id"), nullable=False)
    shipment = relationship("Shipment", back_populates="items")
    
    def __repr__(self):
        return f"<ShipmentItem {self.name}>"

class Document(Base, UUIDMixin, TimestampMixin, BlockchainMixin, BaseCRUD):
    """Documents related to shipments"""
    __tablename__ = "documents"

    name = Column(String, nullable=False)
    document_type = Column(String, nullable=False)  # invoice, bill_of_lading, customs, etc.
    file_path = Column(String, nullable=True)
    content_hash = Column(String, nullable=True)  # SHA-256 hash of the document content
    document_data = Column(JSON, nullable=True)  # Changed from 'metadata' to 'document_data'
    
    # Relationships
    shipments = relationship("Shipment", secondary=shipment_documents, back_populates="documents")
    
    def __repr__(self):
        return f"<Document {self.name}>"