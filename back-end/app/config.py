from pydantic_settings import BaseSettings
from typing import List, Dict, Any, Optional
from functools import lru_cache
import os
from pathlib import Path

class Settings(BaseSettings):
    # Base settings
    APP_NAME: str = "blockchain-logistics-traceability"
    API_V1_STR: str = "/api/v1"
    DEBUG: bool = False
    
    # Security
    SECRET_KEY: str
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 30
    REFRESH_TOKEN_EXPIRE_DAYS: int = 7
    
    # CORS
    CORS_ORIGINS: List[str] = ["*"]
    ALLOW_HOST: List[str] = ["*"]
    
    # Database - both URL format and individual component format
    DATABASE_URL: Optional[str] = None
    ASYNC_DATABASE_URL: Optional[str] = None
    
    # Individual DB connection parameters
    DB_USER: Optional[str] = None
    DB_PASSWORD: Optional[str] = None
    DB_HOST: Optional[str] = None
    DB_PORT: Optional[str] = None
    DB_NAME: Optional[str] = None
    ALLOWED_HOSTS: Optional[str] = None
    
    DB_POOL_SIZE: int = 5
    DB_MAX_OVERFLOW: int = 10
    DB_ECHO: bool = False
    
    # Blockchain settings
    # Ethereum
    BLOCKCHAIN_ETHEREUM_ENABLED: bool = False
    BLOCKCHAIN_ETHEREUM_NODE_URL: Optional[str] = None
    BLOCKCHAIN_ETHEREUM_PRIVATE_KEY: Optional[str] = None
    BLOCKCHAIN_ETHEREUM_SHIPMENT_REGISTRY_ADDRESS: Optional[str] = None
    BLOCKCHAIN_ETHEREUM_EVENT_LOG_ADDRESS: Optional[str] = None
    BLOCKCHAIN_ETHEREUM_CHAIN_ID: int = 1  # Mainnet by default
    BLOCKCHAIN_ETHEREUM_API_KEY: Optional[str] = None
    
    # Substrate
    BLOCKCHAIN_SUBSTRATE_ENABLED: bool = False
    BLOCKCHAIN_SUBSTRATE_NODE_URL: Optional[str] = None
    BLOCKCHAIN_SUBSTRATE_MNEMONIC: Optional[str] = None
    BLOCKCHAIN_SUBSTRATE_SS58_FORMAT: int = 42  # Default for Substrate
    BLOCKCHAIN_SUBSTRATE_API_KEY: Optional[str] = None
    
    # Vietnam Chain (custom L1)
    BLOCKCHAIN_VIETNAMCHAIN_ENABLED: bool = False
    BLOCKCHAIN_VIETNAMCHAIN_NODE_URL: Optional[str] = None
    BLOCKCHAIN_VIETNAMCHAIN_API_KEY: Optional[str] = None
    BLOCKCHAIN_VIETNAMCHAIN_API_SECRET: Optional[str] = None
    BLOCKCHAIN_VIETNAMCHAIN_ORGANIZATION_ID: Optional[str] = None
    
    # Oracle settings
    IOT_API_URL: Optional[str] = None
    IOT_API_KEY: Optional[str] = None
    GPS_API_URL: Optional[str] = None
    GPS_API_KEY: Optional[str] = None
    
    # Bridge settings
    BRIDGE_ENABLED: bool = False
    BRIDGE_ENDPOINTS: Dict[str, str] = {}
    
    # Logging
    LOG_LEVEL: str = "INFO"
    LOG_FORMAT: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    LOG_FILE: Optional[str] = None
    
    # Storage
    STORAGE_PATH: Path = Path("./storage")
    
    # Role definitions
    ROLES: Dict[str, List[str]] = {
        "admin": ["all"],
        "shipper": ["shipment:read", "shipment:create", "event:create", "event:read"],
        "warehouse": ["shipment:read", "event:create", "event:read"],
        "customs": ["shipment:read", "event:read", "alert:read"],
        "client": ["shipment:read", "event:read"],
    }
    
    # Alert thresholds
    ALERT_TIME_THRESHOLD: int = 3600  # seconds
    ALERT_DISTANCE_THRESHOLD: float = 5.0  # km
    
    # For compatibility with alembic
    SQLALCHEMY_DATABASE_URI: str = ""
    
    def __init__(self, **data):
        super().__init__(**data)
        
        # If DATABASE_URL is not provided but individual components are, construct it
        if not self.DATABASE_URL and self.DB_USER and self.DB_PASSWORD and self.DB_HOST and self.DB_NAME:
            port = self.DB_PORT or "5432"
            self.DATABASE_URL = f"postgresql://{self.DB_USER}:{self.DB_PASSWORD}@{self.DB_HOST}:{port}/{self.DB_NAME}"
            
        # If ASYNC_DATABASE_URL is not provided, derive it from DATABASE_URL
        if self.DATABASE_URL and not self.ASYNC_DATABASE_URL:
            self.ASYNC_DATABASE_URL = self.DATABASE_URL.replace("postgresql://", "postgresql+asyncpg://")
            
        # Set SQLALCHEMY_DATABASE_URI for alembic
        if self.DATABASE_URL:
            self.SQLALCHEMY_DATABASE_URI = self.DATABASE_URL
    
    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = True

@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance"""
    return Settings()

# Create instance for importing
settings = get_settings()

# Ensure storage directory exists
os.makedirs(settings.STORAGE_PATH, exist_ok=True)