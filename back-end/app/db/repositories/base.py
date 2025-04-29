from typing import Any, Dict, Generic, List, Optional, Type, TypeVar, Union
from uuid import UUID

from fastapi.encoders import jsonable_encoder
from pydantic import BaseModel
from sqlalchemy import select, update, delete
from sqlalchemy.ext.asyncio import AsyncSession

from app.models.base import BaseCRUD
from app.utils.logger import get_logger

logger = get_logger(__name__)

ModelType = TypeVar("ModelType", bound=BaseCRUD)
CreateSchemaType = TypeVar("CreateSchemaType", bound=BaseModel)
UpdateSchemaType = TypeVar("UpdateSchemaType", bound=BaseModel)


class BaseRepository(Generic[ModelType, CreateSchemaType, UpdateSchemaType]):
    """
    Base repository class for CRUD operations
    
    Attributes:
        model: Model class to operate on
    """
    
    def __init__(self, model: Type[ModelType]):
        self.model = model
    
    async def get(self, db: AsyncSession, id: Union[UUID, str]) -> Optional[ModelType]:
        """
        Get a single record by ID
        
        Args:
            db: Database session
            id: Record ID
            
        Returns:
            Record if found, None otherwise
        """
        query = select(self.model).where(self.model.id == id)
        result = await db.execute(query)
        return result.scalars().first()
    
    async def get_multi(
        self, 
        db: AsyncSession, 
        *, 
        skip: int = 0, 
        limit: int = 100,
        **filters
    ) -> List[ModelType]:
        """
        Get multiple records with optional filtering
        
        Args:
            db: Database session
            skip: Number of records to skip
            limit: Maximum number of records to return
            filters: Additional filtering criteria
            
        Returns:
            List of records
        """
        query = select(self.model)
        
        # Apply filters if any
        for field, value in filters.items():
            if hasattr(self.model, field) and value is not None:
                query = query.where(getattr(self.model, field) == value)
        
        query = query.offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()
    
    async def create(self, db: AsyncSession, *, obj_in: CreateSchemaType) -> ModelType:
        """
        Create a new record
        
        Args:
            db: Database session
            obj_in: Data to create record with
            
        Returns:
            Created record
        """
        obj_in_data = jsonable_encoder(obj_in)
        db_obj = self.model(**obj_in_data)
        db.add(db_obj)
        await db.commit()
        await db.refresh(db_obj)
        return db_obj
    
    async def update(
        self,
        db: AsyncSession,
        *,
        db_obj: ModelType,
        obj_in: Union[UpdateSchemaType, Dict[str, Any]]
    ) -> ModelType:
        """
        Update a record
        
        Args:
            db: Database session
            db_obj: Database object to update
            obj_in: New data to update with
            
        Returns:
            Updated record
        """
        obj_data = jsonable_encoder(db_obj)
        
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
    
    async def delete(self, db: AsyncSession, *, id: Union[UUID, str]) -> Optional[ModelType]:
        """
        Delete a record by ID
        
        Args:
            db: Database session
            id: Record ID
            
        Returns:
            Deleted record or None if not found
        """
        obj = await self.get(db, id=id)
        if obj:
            await db.delete(obj)
            await db.commit()
        return obj
    
    async def count(self, db: AsyncSession, **filters) -> int:
        """
        Count records with optional filtering
        
        Args:
            db: Database session
            filters: Additional filtering criteria
            
        Returns:
            Number of records
        """
        query = select(self.model)
        
        # Apply filters if any
        for field, value in filters.items():
            if hasattr(self.model, field) and value is not None:
                query = query.where(getattr(self.model, field) == value)
        
        result = await db.execute(query)
        return len(result.scalars().all())