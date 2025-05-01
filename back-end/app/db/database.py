# database.py
from sqlalchemy import create_engine
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import NullPool

from app.config import settings

# Create async engine
async_engine = create_async_engine(
    settings.ASYNC_DATABASE_URL,
    pool_pre_ping=True,
    echo=settings.DB_ECHO,
    pool_size=settings.DB_POOL_SIZE,
    max_overflow=settings.DB_MAX_OVERFLOW,
)

# Create sync engine for migrations and admin tasks
sync_engine = create_engine(
    settings.DATABASE_URL,
    pool_pre_ping=True,
    echo=settings.DB_ECHO,
    pool_size=settings.DB_POOL_SIZE,
    max_overflow=settings.DB_MAX_OVERFLOW,
)

# Create async session factory
AsyncSessionLocal = sessionmaker(
    autocommit=False,
    autoflush=False,
    bind=async_engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

# Create sync session factory for migrations and admin tasks
SyncSessionLocal = sessionmaker(
    autocommit=False,
    autoflush=False,
    bind=sync_engine,
)

# Create declarative base for models
Base = declarative_base()

async def init_db() -> None:
    """Initialize database and create all tables"""
    from app.utils.logger import get_logger
    
    logger = get_logger(__name__)
    logger.info("Initializing database...")
    
    # Import all models here inside the function to avoid circular imports
    # This ensures all models are registered with SQLAlchemy's metadata
    from app.models.user import User, Organization
    from app.models.shipment import Shipment, ShipmentItem, Document
    from app.models.event import ShipmentEvent
    from app.models.alert import Alert
    from app.models.blockchain import BlockchainTransaction
    
    async with async_engine.begin() as conn:
        # For development, you may want to drop and recreate all tables
        # Uncomment this if you need to reset the database
        # logger.warning("Dropping all tables... (DEVELOPMENT MODE)")
        # await conn.run_sync(Base.metadata.drop_all)
        
        logger.info("Creating database tables...")
        await conn.run_sync(Base.metadata.create_all)
        logger.info("Database tables created successfully")

async def get_db() -> AsyncSession:
    """Get database session"""
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()