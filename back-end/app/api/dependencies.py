# dependencies.py
from typing import Generator, Optional
from fastapi import Depends, HTTPException, status, Security
from fastapi.security import OAuth2PasswordBearer, APIKeyHeader
from jose import jwt, JWTError
from sqlalchemy.ext.asyncio import AsyncSession
from uuid import UUID

from app.db.database import get_db
from app.core.auth import verify_token
from app.services import auth as auth_service
from app.models.user import User
from app.config import settings

oauth2_scheme = OAuth2PasswordBearer(tokenUrl=f"/api/v1/auth/login")
api_key_header = APIKeyHeader(name="X-API-Key", auto_error=False)

async def get_db() -> Generator[AsyncSession, None, None]:
    async with get_db() as session:
        yield session


async def get_current_user(
    token: str = Depends(oauth2_scheme),
    api_key: Optional[str] = Security(api_key_header),
    db: AsyncSession = Depends(get_db)
) -> User:
    """
    Get the current user from the token or API key
    """
    credentials_exception = HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Invalid authentication credentials",
        headers={"WWW-Authenticate": "Bearer"},
    )
    
    # Try to use token if provided
    if token:
        try:
            payload = verify_token(token)
            user_id: UUID = UUID(payload.get("sub"))
            if user_id is None:
                raise credentials_exception
        except JWTError:
            raise credentials_exception
            
        user = await auth_service.get_user(db, user_id=user_id)
        if user is None:
            raise credentials_exception
        return user
    
    # Try to use API key if provided
    if api_key:
        user = await auth_service.get_user_by_api_key(db, api_key=api_key)
        if user is None:
            raise credentials_exception
        return user
    
    # No valid authentication provided
    raise credentials_exception

async def get_current_active_user(
    current_user: User = Depends(get_current_user),
) -> User:
    """
    Check if the current user is active
    """
    if not current_user.is_active:
        raise HTTPException(status_code=400, detail="Inactive user")
    return current_user

async def get_current_admin_user(
    current_user: User = Depends(get_current_active_user),
) -> User:
    """
    Check if the current user is an admin
    """
    if current_user.role != "admin":
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Insufficient permissions",
        )
    return current_user