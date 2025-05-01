#models/user.py
from sqlalchemy import Column, String, ForeignKey
from sqlalchemy.orm import relationship
from sqlalchemy.dialects.postgresql import UUID

from app.db.database import Base
from app.models.base import TimestampMixin, UUIDMixin, BaseCRUD

class User(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    __tablename__ = "users"

    username = Column(String, unique=True, index=True, nullable=False)
    hashed_password = Column(String, nullable=False)
    full_name = Column(String, nullable=True)
    role = Column(String, default="client", nullable=False)
    
    organization_id = Column(UUID(as_uuid=True), ForeignKey("organizations.id"), nullable=True)
    organization = relationship("Organization", back_populates="users")
    
    def __repr__(self):
        return f"<User {self.username}>"

class Organization(Base, UUIDMixin, TimestampMixin, BaseCRUD):
    __tablename__ = "organizations"

    name = Column(String, nullable=False)
    org_type = Column(String, nullable=False)
    
    users = relationship("User", back_populates="organization")
    shipments = relationship("Shipment", back_populates="organization")
    
    def __repr__(self):
        return f"<Organization {self.name}>"