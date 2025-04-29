# user.py
from sqlalchemy import Boolean, Column, String, ForeignKey, Table
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID
import uuid

from app.db.database import Base
from app.models.base import TimestampMixin, UUIDMixin, BaseCRUD

class User(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """User model for authentication and authorization"""
    __tablename__ = "users"

    email = Column(String, unique=True, index=True, nullable=False)
    hashed_password = Column(String, nullable=False)
    full_name = Column(String, nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    role = Column(String, default="client", nullable=False)  # admin, shipper, warehouse, customs, client
    
    # Organization relationship
    organization_id = Column(UUID(as_uuid=True), ForeignKey("organizations.id"), nullable=True)
    organization = relationship("Organization", back_populates="users")
    
    # API key relationship
    api_keys = relationship("APIKey", back_populates="user", cascade="all, delete-orphan")
    
    # Refresh tokens relationship
    refresh_tokens = relationship("RefreshToken", back_populates="user", cascade="all, delete-orphan")
    
    # Wallet address for blockchain transactions
    wallet_address = Column(String, nullable=True)
    
    def __repr__(self):
        return f"<User {self.email}>"

class Organization(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Organization model for grouping users"""
    __tablename__ = "organizations"

    name = Column(String, nullable=False)
    description = Column(String, nullable=True)
    org_type = Column(String, nullable=False)  # shipper, warehouse, customs, client
    is_active = Column(Boolean, default=True, nullable=False)
    
    # Registration info
    tax_id = Column(String, nullable=True)
    address = Column(String, nullable=True)
    contact_email = Column(String, nullable=True)
    contact_phone = Column(String, nullable=True)
    
    # Relationships
    users = relationship("User", back_populates="organization")
    shipments = relationship("Shipment", back_populates="organization")
    
    def __repr__(self):
        return f"<Organization {self.name}>"

class APIKey(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """API key model for programmatic access"""
    __tablename__ = "api_keys"

    key = Column(String, unique=True, index=True, nullable=False)
    name = Column(String, nullable=False)
    is_active = Column(Boolean, default=True, nullable=False)
    expires_at = Column(Column, nullable=True)
    
    # Relationship to user
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False)
    user = relationship("User", back_populates="api_keys")
    
    def __repr__(self):
        return f"<APIKey {self.name}>"

class RefreshToken(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    """Refresh token model for JWT refresh"""
    __tablename__ = "refresh_tokens"

    token = Column(String, unique=True, index=True, nullable=False)
    expires_at = Column(Column, nullable=False)
    is_revoked = Column(Boolean, default=False, nullable=False)
    
    # Relationship to user
    user_id = Column(UUID(as_uuid=True), ForeignKey("users.id"), nullable=False)
    user = relationship("User", back_populates="refresh_tokens")
    
    def __repr__(self):
        return f"<RefreshToken {self.id}>"