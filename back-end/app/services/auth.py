# app.services.auth.py
from datetime import datetime, timedelta
from typing import Optional, Dict, Any, Tuple
from uuid import UUID
import uuid
import secrets

from fastapi import Depends, HTTPException, status, BackgroundTasks
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.database import get_db
from app.db.repositories import (
    user_repository, 
    organization_repository,
    refresh_token_repository,
    api_key_repository
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
    email: str, 
    password: str
) -> User:
    """
    Authenticate a user with email and password
    
    Args:
        db: Database session
        email: User email
        password: User password
        
    Returns:
        Authenticated user
        
    Raises:
        CredentialsException: If authentication fails
    """
    user = await user_repository.get_by_email(db, email)
    
    if not user:
        logger.warning(f"Authentication failed: User {email} not found")
        raise CredentialsException(detail="Incorrect email or password")
    
    if not verify_password(password, user.hashed_password):
        logger.warning(f"Authentication failed: Invalid password for {email}")
        raise CredentialsException(detail="Incorrect email or password")
    
    if not user.is_active:
        logger.warning(f"Authentication failed: User {email} is inactive")
        raise CredentialsException(detail="Inactive user")
    
    return user

# For FastAPI OAuth2PasswordRequestForm compatibility
async def authenticate_user(
    db: AsyncSession, 
    username: str, 
    password: str
) -> Token:
    """
    Authenticate a user with username/password (username is email)
    
    Args:
        db: Database session
        username: User email
        password: User password
        
    Returns:
        Token with access and refresh tokens
        
    Raises:
        AuthenticationError: If authentication fails
    """
    try:
        user = await get_user_by_email(db, username)
        
        if not user:
            logger.warning(f"Authentication failed: User {username} not found")
            raise AuthenticationError(detail="Incorrect email or password")
        
        if not verify_password(password, user.hashed_password):
            logger.warning(f"Authentication failed: Invalid password for {username}")
            raise AuthenticationError(detail="Incorrect email or password")
        
        if not user.is_active:
            logger.warning(f"Authentication failed: User {username} is inactive")
            raise AuthenticationError(detail="Inactive user")
        
        # Create tokens
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
    """
    Create access and refresh tokens for a user
    
    Args:
        db: Database session
        user_id: User ID
        
    Returns:
        Token object with access and refresh tokens
    """
    # Create access token
    access_token = create_access_token(subject=str(user_id))
    
    # Create refresh token
    refresh_token_value = create_refresh_token(subject=str(user_id))
    
    # Calculate expiry date for refresh token
    expires_at = datetime.utcnow() + timedelta(days=7)  # Match with settings
    
    # Store refresh token in database
    await refresh_token_repository.create(db, obj_in={
        "token": refresh_token_value,
        "expires_at": expires_at,
        "is_revoked": False,
        "user_id": user_id
    })
    
    return Token(
        access_token=access_token,
        refresh_token=refresh_token_value,
        token_type="bearer"
    )

async def refresh_access_token(
    db: AsyncSession, 
    refresh_token: str
) -> Token:
    """
    Refresh access token using a valid refresh token
    
    Args:
        db: Database session
        refresh_token: Refresh token
        
    Returns:
        New token pair
        
    Raises:
        CredentialsException: If refresh token is invalid
    """
    # Get the refresh token from database
    token_data = await refresh_token_repository.get_by_token(db, refresh_token)
    
    if not token_data:
        logger.warning("Token refresh failed: Token not found")
        raise CredentialsException(detail="Invalid refresh token")
    
    # Check if token is expired
    if token_data.expires_at < datetime.utcnow():
        logger.warning("Token refresh failed: Token expired")
        raise CredentialsException(detail="Refresh token expired")
    
    # Check if token is revoked
    if token_data.is_revoked:
        logger.warning("Token refresh failed: Token revoked")
        raise CredentialsException(detail="Refresh token revoked")
    
    # Revoke the current refresh token
    await refresh_token_repository.revoke_token(db, refresh_token)
    
    # Create new tokens
    return await create_user_tokens(db, token_data.user_id)

async def refresh_token(
    db: AsyncSession, 
    refresh_token: str
) -> Token:
    """
    Refresh access token using a valid refresh token (alias for refresh_access_token)
    
    Args:
        db: Database session
        refresh_token: Refresh token
        
    Returns:
        New token pair
        
    Raises:
        AuthenticationError: If refresh token is invalid
    """
    try:
        return await refresh_access_token(db, refresh_token)
    except CredentialsException as e:
        logger.warning(f"Token refresh failed: {e.detail}")
        raise AuthenticationError(detail=e.detail)

async def get_user(
    db: AsyncSession, 
    user_id: UUID
) -> Optional[User]:
    """
    Get a user by ID
    
    Args:
        db: Database session
        user_id: User ID
        
    Returns:
        User if found, None otherwise
    """
    return await user_repository.get(db, id=user_id)

async def get_user_by_id(id: UUID, db: AsyncSession = Depends(get_db)) -> Optional[User]:
    """
    Get a user by ID
    
    Args:
        id: User ID
        db: Database session
        
    Returns:
        User if found, None otherwise
    """
    return await user_repository.get(db, id=id)

async def get_user_by_email(email: str, db: AsyncSession = Depends(get_db)) -> Optional[User]:
    """
    Get a user by email
    
    Args:
        email: User email
        db: Database session
        
    Returns:
        User if found, None otherwise
    """
    return await user_repository.get_by_email(db, email)

async def get_user_by_api_key(
    db: AsyncSession, 
    api_key: str
) -> Optional[User]:
    """
    Get a user by API key
    
    Args:
        db: Database session
        api_key: API key
        
    Returns:
        User if found, None otherwise
    """
    api_key_obj = await api_key_repository.get_by_key(db, api_key)
    if not api_key_obj:
        return None
    return await user_repository.get(db, id=api_key_obj.user_id)

async def create_user(
    db: AsyncSession, 
    user_in: UserCreate, 
    is_superuser: bool = False
) -> User:
    """
    Create a new user
    
    Args:
        db: Database session
        user_in: User data
        is_superuser: Whether to create a superuser
        
    Returns:
        Created user
        
    Raises:
        ResourceAlreadyExistsError: If user with same email already exists
    """
    # Check if user with same email exists
    existing_user = await user_repository.get_by_email(db, user_in.email)
    if existing_user:
        logger.warning(f"User creation failed: Email {user_in.email} already exists")
        raise ResourceAlreadyExistsError(detail="User with this email already exists")
    
    # Set role to admin if superuser
    if is_superuser:
        user_in.role = "admin"
    
    # Create the user
    user = await user_repository.create(db, obj_in=user_in)
    logger.info(f"User created: {user.email} (ID: {user.id})")
    
    return user

async def register_user(
    db: AsyncSession,
    user_in: UserCreate,
    background_tasks: Optional[BackgroundTasks] = None
) -> User:
    """
    Register a new user
    
    Args:
        db: Database session
        user_in: User data
        background_tasks: Background tasks runner for sending emails, etc.
        
    Returns:
        Created user
        
    Raises:
        ResourceAlreadyExistsError: If user with same email already exists
        ResourceNotFoundError: If organization_id is provided but doesn't exist
    """
    # Check if user email already exists
    existing_user = await user_repository.get_by_email(db, user_in.email)
    if existing_user:
        logger.warning(f"User registration failed: Email {user_in.email} already exists")
        raise ResourceAlreadyExistsError(detail="User with this email already exists")
    
    # If organization_id is provided, validate it exists
    if user_in.organization_id:
        organization = await organization_repository.get(db, id=user_in.organization_id)
        if not organization:
            logger.warning(f"User registration failed: Organization {user_in.organization_id} not found")
            # Set organization_id to None to avoid foreign key constraint violation
            user_in_dict = user_in.dict()
            user_in_dict["organization_id"] = None
            
            # Create a new UserCreate instance with the modified dict
            from app.schemas.auth import UserCreate
            user_in = UserCreate(**user_in_dict)
            
            logger.info(f"Setting organization_id to None for user {user_in.email} as the organization doesn't exist")
    
    # Create the user
    user = await create_user(db, user_in)
    
    # Send welcome email in background if tasks runner provided
    if background_tasks:
        # This would be implemented separately in a notifications service
        # background_tasks.add_task(send_welcome_email, user.email, user.full_name)
        pass
    
    logger.info(f"User registered successfully: {user.email} (ID: {user.id})")
    return user

async def update_user(
    db: AsyncSession, 
    user_id: UUID, 
    user_in: UserUpdate
) -> User:
    """
    Update a user
    
    Args:
        db: Database session
        user_id: User ID
        user_in: New user data
        
    Returns:
        Updated user
        
    Raises:
        ResourceNotFoundError: If user not found
    """
    # Get the user
    user = await user_repository.get(db, id=user_id)
    if not user:
        logger.warning(f"User update failed: User {user_id} not found")
        raise ResourceNotFoundError(detail="User not found")
    
    # Handle password update separately if provided
    if user_in.password:
        # Update will handle the password hashing
        user_in_dict = user_in.dict(exclude_unset=True)
        if "password" in user_in_dict:
            user_in_dict["hashed_password"] = get_password_hash(user_in_dict.pop("password"))
        
        # Update user
        user = await user_repository.update(db, db_obj=user, obj_in=user_in_dict)
    else:
        # Update user without password change
        user = await user_repository.update(db, db_obj=user, obj_in=user_in)
    
    logger.info(f"User updated: {user.email} (ID: {user.id})")
    return user

async def create_organization(
    db: AsyncSession, 
    organization_in: OrganizationCreate
) -> Organization:
    """
    Create a new organization
    
    Args:
        db: Database session
        organization_in: Organization data
        
    Returns:
        Created organization
        
    Raises:
        ResourceAlreadyExistsError: If organization with same name already exists
    """
    # Check if organization with same name exists
    existing_org = await organization_repository.get_by_name(db, organization_in.name)
    if existing_org:
        logger.warning(f"Organization creation failed: Name {organization_in.name} already exists")
        raise ResourceAlreadyExistsError(detail="Organization with this name already exists")
    
    # Create the organization
    organization = await organization_repository.create(db, obj_in=organization_in)
    logger.info(f"Organization created: {organization.name} (ID: {organization.id})")
    
    return organization

async def revoke_all_user_sessions(db: AsyncSession, user_id: UUID) -> int:
    """
    Revoke all refresh tokens for a user
    
    Args:
        db: Database session
        user_id: User ID
        
    Returns:
        Number of tokens revoked
    """
    count = await refresh_token_repository.revoke_all_user_tokens(db, user_id)
    logger.info(f"Revoked {count} tokens for user {user_id}")
    return count

async def logout(
    db: AsyncSession,
    user_id: UUID,
    refresh_token: str
) -> bool:
    """
    Logout a user by revoking their refresh token
    
    Args:
        db: Database session
        user_id: User ID
        refresh_token: Refresh token to revoke
        
    Returns:
        True if logout successful, False otherwise
    """
    try:
        # Get the token from database
        token_data = await refresh_token_repository.get_by_token(db, refresh_token)
        
        # If token not found or belongs to another user, fail
        if not token_data or token_data.user_id != user_id:
            logger.warning(f"Logout failed: Token not found or invalid for user {user_id}")
            return False
        
        # Revoke the token
        await refresh_token_repository.revoke_token(db, refresh_token)
        logger.info(f"User logged out: {user_id}")
        
        return True
    
    except Exception as e:
        logger.error(f"Logout error: {str(e)}")
        return False

async def generate_api_key(
    db: AsyncSession,
    user_id: UUID,
    name: str = "Default API Key"
) -> str:
    """
    Generate a new API key for a user
    
    Args:
        db: Database session
        user_id: User ID
        name: Name of the API key
        
    Returns:
        Generated API key
    """
    # Check if user exists
    user = await user_repository.get(db, id=user_id)
    if not user:
        logger.warning(f"API key generation failed: User {user_id} not found")
        raise ResourceNotFoundError(detail="User not found")
    
    # Generate a secure random API key
    api_key = secrets.token_urlsafe(32)
    
    # Create API key record
    await api_key_repository.create(db, obj_in={
        "key": api_key,
        "name": name,
        "user_id": user_id,
        "is_active": True,
        "expires_at": None  # No expiration by default
    })
    
    logger.info(f"API key generated for user: {user_id}")
    return api_key