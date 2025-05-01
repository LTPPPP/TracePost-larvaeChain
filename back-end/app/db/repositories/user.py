# db/repositories/user.py
from typing import Optional, List
from uuid import UUID
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.repositories.base import BaseRepository
from app.models.user import User, Organization
from app.schemas.auth import UserCreate, UserUpdate, OrganizationCreate, OrganizationUpdate
from app.core.security import get_password_hash
from app.utils.logger import get_logger

logger = get_logger(__name__)

class UserRepository(BaseRepository[User, UserCreate, UserUpdate]):
    def __init__(self):
        super().__init__(User)
    
    async def get_by_username(self, db: AsyncSession, username: str) -> Optional[User]:
        query = select(User).where(User.username == username)
        result = await db.execute(query)
        return result.scalars().first()
    
    async def create(self, db: AsyncSession, *, obj_in: UserCreate) -> User:
        db_obj = User(
            username=obj_in.username,
            hashed_password=get_password_hash(obj_in.password),
            full_name=obj_in.full_name,
            organization_id=obj_in.organization_id,
            role=obj_in.role
        )
        db.add(db_obj)
        await db.commit()
        await db.refresh(db_obj)
        return db_obj
    
    async def update_password(
        self, 
        db: AsyncSession, 
        *, 
        user_id: UUID, 
        password: str
    ) -> Optional[User]:
        user = await self.get(db, id=user_id)
        if not user:
            return None
            
        user.hashed_password = get_password_hash(password)
        db.add(user)
        await db.commit()
        await db.refresh(user)
        return user

    async def get_users_by_organization(
        self,
        db: AsyncSession,
        *,
        organization_id: UUID,
        skip: int = 0,
        limit: int = 100
    ) -> List[User]:
        query = select(User).where(User.organization_id == organization_id)
        query = query.offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()


class OrganizationRepository(BaseRepository[Organization, OrganizationCreate, OrganizationUpdate]):
    def __init__(self):
        super().__init__(Organization)
    
    async def get_by_name(self, db: AsyncSession, name: str) -> Optional[Organization]:
        query = select(Organization).where(Organization.name == name)
        result = await db.execute(query)
        return result.scalars().first()