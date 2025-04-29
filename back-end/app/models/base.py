# base.py
from sqlalchemy import Column, DateTime, String, func
from sqlalchemy.dialects.postgresql import UUID
import uuid

from app.db.database import Base

class TimestampMixin:
    """Mixin that adds timestamp fields to models"""
    created_at = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    updated_at = Column(DateTime(timezone=True), server_default=func.now(), onupdate=func.now(), nullable=False)

class UUIDMixin:
    """Mixin that adds UUID primary key to models"""
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4, index=True)

class BlockchainMixin:
    """Mixin that adds blockchain transaction fields to models"""
    blockchain_tx_hash = Column(String, index=True, nullable=True)
    blockchain_network = Column(String, nullable=True)
    blockchain_timestamp = Column(DateTime(timezone=True), nullable=True)
    blockchain_status = Column(String, default="pending", nullable=False)  # pending, confirmed, failed

class BaseCRUD:
    """Base class for CRUD operations"""
    @classmethod
    async def get(cls, db, id):
        """Get a single object by ID"""
        return await db.get(cls, id)

    @classmethod
    async def get_multi(cls, db, *, skip=0, limit=100):
        """Get multiple objects"""
        query = db.query(cls).offset(skip).limit(limit)
        return await query.all()

    @classmethod
    async def create(cls, db, *, obj_in):
        """Create a new object"""
        obj_in_data = obj_in.dict() if hasattr(obj_in, "dict") else obj_in
        db_obj = cls(**obj_in_data)
        db.add(db_obj)
        await db.commit()
        await db.refresh(db_obj)
        return db_obj

    @classmethod
    async def update(cls, db, *, db_obj, obj_in):
        """Update an object"""
        obj_data = db_obj.__dict__
        if isinstance(obj_in, dict):
            update_data = obj_in
        else:
            update_data = obj_in.dict(exclude_unset=True)
        for field in obj_data:
            if field in update_data:
                setattr(db_obj, field, update_data[field])
        db.add(db_obj)
        await db.commit()
        await db.refresh(db_obj)
        return db_obj

    @classmethod
    async def delete(cls, db, *, id):
        """Delete an object"""
        obj = await db.get(cls, id)
        await db.delete(obj)
        await db.commit()
        return obj