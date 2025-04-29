from typing import Optional, List
from uuid import UUID
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.repositories.base import BaseRepository
from app.models.user import User, Organization, APIKey, RefreshToken
from app.schemas.auth import UserCreate, UserUpdate, OrganizationCreate, OrganizationUpdate
from app.core.security import get_password_hash
from app.utils.logger import get_logger

logger = get_logger(__name__)

class UserRepository(BaseRepository[User, UserCreate, UserUpdate]):
    """Repository for User operations"""
    
    def __init__(self):
        super().__init__(User)
    
    async def get_by_email(self, db: AsyncSession, email: str) -> Optional[User]:
        """
        Get a user by email
        
        Args:
            db: Database session
            email: User email
            
        Returns:
            User if found, None otherwise
        """
        query = select(User).where(User.email == email)
        result = await db.execute(query)
        return result.scalars().first()
    
    async def get_by_wallet_address(self, db: AsyncSession, wallet_address: str) -> Optional[User]:
        """
        Get a user by wallet address
        
        Args:
            db: Database session
            wallet_address: Wallet address
            
        Returns:
            User if found, None otherwise
        """
        query = select(User).where(User.wallet_address == wallet_address)
        result = await db.execute(query)
        return result.scalars().first()
    
    async def create(self, db: AsyncSession, *, obj_in: UserCreate) -> User:
        """
        Create a new user with hashed password
        
        Args:
            db: Database session
            obj_in: User data
            
        Returns:
            Created user
        """
        db_obj = User(
            email=obj_in.email,
            hashed_password=get_password_hash(obj_in.password),
            full_name=obj_in.full_name,
            organization_id=obj_in.organization_id,
            role=obj_in.role,
            is_active=True
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
        """
        Update user password
        
        Args:
            db: Database session
            user_id: User ID
            password: New password
            
        Returns:
            Updated user or None if not found
        """
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
        """
        Get users by organization
        
        Args:
            db: Database session
            organization_id: Organization ID
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of users
        """
        query = select(User).where(User.organization_id == organization_id)
        query = query.offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()


class OrganizationRepository(BaseRepository[Organization, OrganizationCreate, OrganizationUpdate]):
    """Repository for Organization operations"""
    
    def __init__(self):
        super().__init__(Organization)
    
    async def get_by_name(self, db: AsyncSession, name: str) -> Optional[Organization]:
        """
        Get an organization by name
        
        Args:
            db: Database session
            name: Organization name
            
        Returns:
            Organization if found, None otherwise
        """
        query = select(Organization).where(Organization.name == name)
        result = await db.execute(query)
        return result.scalars().first()


class APIKeyRepository(BaseRepository):
    """Repository for API key operations"""
    
    def __init__(self):
        super().__init__(APIKey)
    
    async def get_by_key(self, db: AsyncSession, key: str) -> Optional[APIKey]:
        """
        Get an API key by key value
        
        Args:
            db: Database session
            key: API key
            
        Returns:
            APIKey if found, None otherwise
        """
        query = select(APIKey).where(APIKey.key == key, APIKey.is_active == True)
        result = await db.execute(query)
        return result.scalars().first()
    
    async def get_by_user(
        self,
        db: AsyncSession,
        *,
        user_id: UUID,
        skip: int = 0,
        limit: int = 100
    ) -> List[APIKey]:
        """
        Get API keys by user
        
        Args:
            db: Database session
            user_id: User ID
            skip: Number of records to skip
            limit: Maximum number of records to return
            
        Returns:
            List of API keys
        """
        query = select(APIKey).where(APIKey.user_id == user_id)
        query = query.offset(skip).limit(limit)
        result = await db.execute(query)
        return result.scalars().all()


class RefreshTokenRepository(BaseRepository):
    """Repository for refresh token operations"""
    
    def __init__(self):
        super().__init__(RefreshToken)
    
    async def get_by_token(self, db: AsyncSession, token: str) -> Optional[RefreshToken]:
        """
        Get a refresh token by token value
        
        Args:
            db: Database session
            token: Refresh token
            
        Returns:
            RefreshToken if found, None otherwise
        """
        query = select(RefreshToken).where(
            RefreshToken.token == token, 
            RefreshToken.is_revoked == False
        )
        result = await db.execute(query)
        return result.scalars().first()
    
    async def revoke_token(self, db: AsyncSession, token: str) -> Optional[RefreshToken]:
        """
        Revoke a refresh token
        
        Args:
            db: Database session
            token: Refresh token
            
        Returns:
            Updated RefreshToken or None if not found
        """
        refresh_token = await self.get_by_token(db, token)
        if not refresh_token:
            return None
            
        refresh_token.is_revoked = True
        db.add(refresh_token)
        await db.commit()
        await db.refresh(refresh_token)
        return refresh_token
    
    async def revoke_all_user_tokens(self, db: AsyncSession, user_id: UUID) -> int:
        """
        Revoke all refresh tokens for a user
        
        Args:
            db: Database session
            user_id: User ID
            
        Returns:
            Number of tokens revoked
        """
        query = select(RefreshToken).where(
            RefreshToken.user_id == user_id,
            RefreshToken.is_revoked == False
        )
        result = await db.execute(query)
        tokens = result.scalars().all()
        
        for token in tokens:
            token.is_revoked = True
            db.add(token)
        
        if tokens:
            await db.commit()
            
        return len(tokens)