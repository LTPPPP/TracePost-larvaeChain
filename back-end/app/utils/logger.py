import logging
import sys
import os
from logging.handlers import RotatingFileHandler
from pathlib import Path

from app.config import settings

def setup_logging():
    """
    Configure logging for the application
    """
    # Create logger
    logger = logging.getLogger()
    logger.setLevel(getattr(logging, settings.LOG_LEVEL))
    
    # Create formatter
    formatter = logging.Formatter(settings.LOG_FORMAT)
    
    # Create console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setFormatter(formatter)
    logger.addHandler(console_handler)
    
    # Create file handler if LOG_FILE is set
    if settings.LOG_FILE:
        log_path = Path(settings.LOG_FILE).resolve()
        os.makedirs(log_path.parent, exist_ok=True)
        
        file_handler = RotatingFileHandler(
            log_path,
            maxBytes=10 * 1024 * 1024,  # 10 MB
            backupCount=5
        )
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)
    
    # Disable other loggers
    for log_name in ["uvicorn", "uvicorn.access"]:
        uvicorn_logger = logging.getLogger(log_name)
        uvicorn_logger.propagate = False
        
        # Add handlers to uvicorn loggers
        for handler in logger.handlers:
            uvicorn_logger.addHandler(handler)
    
    return logger

def get_logger(name):
    """
    Get a logger with the given name
    
    Args:
        name: Logger name
        
    Returns:
        Logger instance
    """
    return logging.getLogger(name)