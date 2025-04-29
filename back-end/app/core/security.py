# security.py
from datetime import datetime, timedelta
from typing import Any, Dict, Optional, Union

from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jose import jwt, JWTError
from passlib.context import CryptContext
from pydantic import ValidationError

from app.config import settings
from app.core.exceptions import CredentialsException, InactiveUserException
from app.models.user import User
from app.schemas.auth import TokenPayload

# Password hashing context
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

# OAuth2 scheme for token authentication
oauth2_scheme = OAuth2PasswordBearer(tokenUrl=f"{settings.API_V1_STR}/auth/login")

def create_access_token(subject: Union[str, Any], expires_delta: Optional[timedelta] = None) -> str:
    """
    Create a JWT access token
    
    Args:
        subject: Subject of the token (usually user ID)
        expires_delta: Token expiration time
        
    Returns:
        Encoded JWT token
    """
    if expires_delta:
        expire = datetime.utcnow() + expires_delta
    else:
        expire = datetime.utcnow() + timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)
    
    to_encode = {"exp": expire, "sub": str(subject), "type": "access"}
    encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
    return encoded_jwt

def create_refresh_token(subject: Union[str, Any]) -> str:
    """
    Create a JWT refresh token
    
    Args:
        subject: Subject of the token (usually user ID)
        
    Returns:
        Encoded JWT token
    """
    expire = datetime.utcnow() + timedelta(days=settings.REFRESH_TOKEN_EXPIRE_DAYS)
    to_encode = {"exp": expire, "sub": str(subject), "type": "refresh"}
    encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
    return encoded_jwt

def verify_password(plain_password: str, hashed_password: str) -> bool:
    """
    Verify password against hash
    
    Args:
        plain_password: Plain text password
        hashed_password: Hashed password
        
    Returns:
        True if password matches hash
    """
    return pwd_context.verify(plain_password, hashed_password)

def get_password_hash(password: str) -> str:
    """
    Hash a password
    
    Args:
        password: Plain text password
        
    Returns:
        Hashed password
    """
    return pwd_context.hash(password)

async def get_current_user(token: str = Depends(oauth2_scheme)) -> User:
    """
    Get the current user from the token
    
    Args:
        token: JWT token
        
    Returns:
        User object
    
    Raises:
        CredentialsException: If token is invalid
    """
    try:
        payload = jwt.decode(
            token, settings.SECRET_KEY, algorithms=[settings.ALGORITHM]
        )
        token_data = TokenPayload(**payload)
        
        # Check token expiration
        if token_data.exp < datetime.utcnow().timestamp():
            raise CredentialsException(detail="Token expired")
        
        # Check token type
        if token_data.type != "access":
            raise CredentialsException(detail="Invalid token type")
        
    except (JWTError, ValidationError):
        raise CredentialsException()
    
    from app.services.auth import get_user_by_id
    user = await get_user_by_id(token_data.sub)
    
    if not user:
        raise CredentialsException()
    
    return user

async def get_current_active_user(current_user: User = Depends(get_current_user)) -> User:
    """
    Get current active user
    
    Args:
        current_user: Current user object
        
    Returns:
        User object if active
        
    Raises:
        InactiveUserException: If user is inactive
    """
    if not current_user.is_active:
        raise InactiveUserException()
    
    return current_user

def has_permission(user: User, required_permission: str) -> bool:
    """
    Check if user has a specific permission
    
    Args:
        user: User object
        required_permission: Permission to check
        
    Returns:
        True if user has permission
    """
    # Get role permissions from settings
    role_permissions = settings.ROLES.get(user.role, [])
    
    # Admin has all permissions
    if "all" in role_permissions:
        return True
    
    # Check specific permission
    return required_permission in role_permissions

def permission_required(required_permission: str):
    """
    Dependency to check if user has required permission
    
    Args:
        required_permission: Permission to check
        
    Returns:
        Dependency function
    """
    async def _permission_required(current_user: User = Depends(get_current_active_user)):
        if not has_permission(current_user, required_permission):
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail=f"Permission denied: {required_permission} required",
            )
        return current_user
    
    return _permission_required