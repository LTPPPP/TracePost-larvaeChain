from datetime import datetime, timedelta
from typing import Optional, Dict, Any, Tuple
from uuid import UUID
import uuid

from fastapi import Depends, HTTPException, status, BackgroundTasks
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.database import get_db
from app.db.repositories import (
    user_repository, 
    organization_repository
)
from app.models.user import User, Organization
from app.schemas.auth import (
    UserCreate, 
    UserUpdate, 
    TokenPayload, 
    Token,
    OrganizationCreate,
    LoginRequest
)
from app.core.security import (
    verify_password, 
    get_password_hash,
    create_access_token,
    create_refresh_token
)
from app.core.exceptions import (
    CredentialsException,
    ResourceNotFoundError,
    ResourceAlreadyExistsError,
    AuthenticationError
)
from app.utils.logger import get_logger

logger = get_logger(__name__)

async def authenticate_user(
    db: AsyncSession, 
    username: str, 
    password: str
) -> Token:
    try:
        user = await user_repository.get_by_username(db, username)
            
        if not user:
            logger.warning(f"Authentication failed: User {username} not found")
            raise AuthenticationError(detail="Incorrect username or password")
        
        if not verify_password(password, user.hashed_password):
            logger.warning(f"Authentication failed: Invalid password for {username}")
            raise AuthenticationError(detail="Incorrect username or password")
        
        return await create_user_tokens(db, user.id)
        
    except Exception as e:
        if not isinstance(e, AuthenticationError):
            logger.error(f"Authentication error: {str(e)}")
            raise AuthenticationError(detail="Authentication failed")
        raise

async def create_user_tokens(
    db: AsyncSession, 
    user_id: UUID
) -> Token:
    access_token = create_access_token(subject=str(user_id))
    refresh_token_value = create_refresh_token(subject=str(user_id))
    
    return Token(
        access_token=access_token,
        refresh_token=refresh_token_value,
        token_type="bearer"
    )

async def get_user(
    db: AsyncSession, 
    user_id: UUID
) -> Optional[User]:
    return await user_repository.get(db, id=user_id)

async def get_user_by_id(id: UUID, db: AsyncSession = Depends(get_db)) -> Optional[User]:
    return await user_repository.get(db, id=id)

async def get_user_by_username(username: str, db: AsyncSession = Depends(get_db)) -> Optional[User]:
    return await user_repository.get_by_username(db, username)

async def create_user(
    db: AsyncSession, 
    user_in: UserCreate, 
    is_superuser: bool = False
) -> User:
    existing_user = await user_repository.get_by_username(db, user_in.username)
    if existing_user:
        logger.warning(f"User creation failed: Username {user_in.username} already exists")
        raise ResourceAlreadyExistsError(detail="User with this username already exists")
    
    if is_superuser:
        user_in.role = "admin"
    
    user = await user_repository.create(db, obj_in=user_in)
    logger.info(f"User created: {user.username} (ID: {user.id})")
    
    return user

async def register_user(
    db: AsyncSession,
    user_in: UserCreate,
    background_tasks: Optional[BackgroundTasks] = None
) -> User:
    existing_user = await user_repository.get_by_username(db, user_in.username)
    if existing_user:
        logger.warning(f"User registration failed: Username {user_in.username} already exists")
        raise ResourceAlreadyExistsError(detail="User with this username already exists")
    
    if user_in.organization_id:
        organization = await organization_repository.get(db, id=user_in.organization_id)
        if not organization:
            logger.warning(f"User registration failed: Organization {user_in.organization_id} not found")
            user_in_dict = user_in.dict()
            user_in_dict["organization_id"] = None
            
            from app.schemas.auth import UserCreate
            user_in = UserCreate(**user_in_dict)
            
            logger.info(f"Setting organization_id to None for user {user_in.username} as the organization doesn't exist")
    
    user = await create_user(db, user_in)
    
    logger.info(f"User registered successfully: {user.username} (ID: {user.id})")
    return user

async def update_user(
    db: AsyncSession, 
    user_id: UUID, 
    user_in: UserUpdate
) -> User:
    user = await user_repository.get(db, id=user_id)
    if not user:
        logger.warning(f"User update failed: User {user_id} not found")
        raise ResourceNotFoundError(detail="User not found")
    
    if user_in.password:
        user_in_dict = user_in.dict(exclude_unset=True)
        if "password" in user_in_dict:
            user_in_dict["hashed_password"] = get_password_hash(user_in_dict.pop("password"))
        
        user = await user_repository.update(db, db_obj=user, obj_in=user_in_dict)
    else:
        user = await user_repository.update(db, db_obj=user, obj_in=user_in)
    
    logger.info(f"User updated: {user.username} (ID: {user.id})")
    return user

async def create_organization(
    db: AsyncSession, 
    organization_in: OrganizationCreate
) -> Organization:
    existing_org = await organization_repository.get_by_name(db, organization_in.name)
    if existing_org:
        logger.warning(f"Organization creation failed: Name {organization_in.name} already exists")
        raise ResourceAlreadyExistsError(detail="Organization with this name already exists")
    
    organization = await organization_repository.create(db, obj_in=organization_in)
    logger.info(f"Organization created: {organization.name} (ID: {organization.id})")
    
    return organization