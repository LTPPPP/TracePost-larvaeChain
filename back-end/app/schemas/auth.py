# auth.py
from pydantic import BaseModel, EmailStr, Field, UUID4, validator
from typing import Optional, List, Dict, Any
from datetime import datetime

class UserBase(BaseModel):
    """Base schema for user data"""
    email: EmailStr
    full_name: Optional[str] = None
    organization_id: Optional[UUID4] = None
    role: Optional[str] = "client"

class UserCreate(UserBase):
    """Schema for creating a new user"""
    password: str = Field(..., min_length=8)
    
    @validator("password")
    def validate_password(cls, v):
        """Validate password complexity"""
        if len(v) < 8:
            raise ValueError("Password must be at least 8 characters long")
        # Add more password validation logic as needed
        return v

class UserUpdate(BaseModel):
    """Schema for updating a user"""
    email: Optional[EmailStr] = None
    full_name: Optional[str] = None
    password: Optional[str] = None
    is_active: Optional[bool] = None
    role: Optional[str] = None
    organization_id: Optional[UUID4] = None
    wallet_address: Optional[str] = None

class UserInDBBase(UserBase):
    """Base schema for user in DB"""
    id: UUID4
    is_active: bool = True
    created_at: datetime
    updated_at: datetime
    wallet_address: Optional[str] = None
    
    class Config:
        orm_mode = True

class User(UserInDBBase):
    """Schema for user response"""
    pass

class UserInDB(UserInDBBase):
    """Schema for user in DB with hashed password"""
    hashed_password: str

class Token(BaseModel):
    """Schema for authentication tokens"""
    access_token: str
    refresh_token: str
    token_type: str = "bearer"

class TokenPayload(BaseModel):
    """Schema for JWT token payload"""
    sub: str
    exp: float
    type: str  # "access" or "refresh"

class LoginRequest(BaseModel):
    """Schema for login request"""
    email: EmailStr
    password: str

class RefreshTokenRequest(BaseModel):
    """Schema for refresh token request"""
    refresh_token: str

class OrganizationBase(BaseModel):
    """Base schema for organization data"""
    name: str
    description: Optional[str] = None
    org_type: str
    tax_id: Optional[str] = None
    address: Optional[str] = None
    contact_email: Optional[EmailStr] = None
    contact_phone: Optional[str] = None

class OrganizationCreate(OrganizationBase):
    """Schema for creating a new organization"""
    pass

class OrganizationUpdate(BaseModel):
    """Schema for updating an organization"""
    name: Optional[str] = None
    description: Optional[str] = None
    org_type: Optional[str] = None
    is_active: Optional[bool] = None
    tax_id: Optional[str] = None
    address: Optional[str] = None
    contact_email: Optional[EmailStr] = None
    contact_phone: Optional[str] = None

class OrganizationInDBBase(OrganizationBase):
    """Base schema for organization in DB"""
    id: UUID4
    is_active: bool = True
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True

class Organization(OrganizationInDBBase):
    """Schema for organization response"""
    pass

class APIKeyBase(BaseModel):
    """Base schema for API key data"""
    name: str
    expires_at: Optional[datetime] = None

class APIKeyCreate(APIKeyBase):
    """Schema for creating a new API key"""
    user_id: UUID4

class APIKeyResponse(APIKeyBase):
    """Schema for API key response"""
    id: UUID4
    key: str  # Only returned once upon creation
    is_active: bool = True
    created_at: datetime
    
    class Config:
        orm_mode = True

class APIKeyInDB(APIKeyBase):
    """Schema for API key in DB"""
    id: UUID4
    key: str
    user_id: UUID4
    is_active: bool = True
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True

class RefreshTokenBase(BaseModel):
    """Base schema for refresh token data"""
    token: str
    expires_at: datetime
    is_revoked: bool = False
    user_id: UUID4

class RefreshTokenCreate(RefreshTokenBase):
    """Schema for creating a new refresh token"""
    pass

class RefreshTokenInDB(RefreshTokenBase):
    """Schema for refresh token in DB"""
    id: UUID4
    created_at: datetime
    updated_at: datetime
    
    class Config:
        orm_mode = True