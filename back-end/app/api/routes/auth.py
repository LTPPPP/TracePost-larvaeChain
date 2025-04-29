# auth.py
from typing import Dict, Any
from fastapi import APIRouter, Depends, HTTPException, status, BackgroundTasks
from fastapi.security import OAuth2PasswordRequestForm
from sqlalchemy.ext.asyncio import AsyncSession

from app.api.dependencies import get_db, get_current_user, get_current_active_user
from app.services import auth as auth_service
from app.models.user import User
from app.schemas.auth import UserCreate, UserResponse, TokenResponse, TokenData, UserUpdate
from app.core.exceptions import AuthenticationError, ResourceNotFoundError

router = APIRouter(prefix="/auth", tags=["Authentication"])

@router.post("/login", response_model=TokenResponse)
async def login_user(
    form_data: OAuth2PasswordRequestForm = Depends(),
    db: AsyncSession = Depends(get_db)
):
    """
    Authenticate a user and return access token
    """
    try:
        tokens = await auth_service.authenticate_user(
            db=db,
            username=form_data.username,
            password=form_data.password
        )
        return tokens
    except AuthenticationError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=str(e),
            headers={"WWW-Authenticate": "Bearer"},
        )

@router.post("/refresh", response_model=TokenResponse)
async def refresh_token(
    token_data: TokenData,
    db: AsyncSession = Depends(get_db)
):
    """
    Refresh an access token using a refresh token
    """
    try:
        tokens = await auth_service.refresh_token(
            db=db,
            refresh_token=token_data.refresh_token
        )
        return tokens
    except AuthenticationError as e:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=str(e),
            headers={"WWW-Authenticate": "Bearer"},
        )

@router.post("/register", response_model=UserResponse)
async def register_user(
    user_in: UserCreate,
    background_tasks: BackgroundTasks,
    db: AsyncSession = Depends(get_db)
):
    """
    Register a new user
    """
    try:
        user = await auth_service.register_user(
            db=db,
            user_in=user_in,
            background_tasks=background_tasks
        )
        return user
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )

@router.get("/me", response_model=UserResponse)
async def get_user_me(
    current_user: User = Depends(get_current_active_user)
):
    """
    Get current user information
    """
    return current_user

@router.put("/me", response_model=UserResponse)
async def update_user_me(
    user_in: UserUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Update current user information
    """
    try:
        user = await auth_service.update_user(
            db=db,
            user_id=current_user.id,
            obj_in=user_in
        )
        return user
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e)
        )

@router.post("/logout", response_model=Dict[str, Any])
async def logout(
    token_data: TokenData,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Logout and invalidate the refresh token
    """
    try:
        success = await auth_service.logout(
            db=db,
            user_id=current_user.id,
            refresh_token=token_data.refresh_token
        )
        return {"success": success, "message": "User logged out successfully"}
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )

@router.post("/api-key", response_model=Dict[str, Any])
async def generate_api_key(
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    """
    Generate a new API key for the current user
    """
    try:
        api_key = await auth_service.generate_api_key(
            db=db,
            user_id=current_user.id
        )
        return {"api_key": api_key}
    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=str(e)
        )