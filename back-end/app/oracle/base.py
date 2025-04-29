from abc import ABC, abstractmethod
from typing import Dict, Any, Optional, List
import asyncio
import json
from datetime import datetime

from app.utils.logger import get_logger

logger = get_logger(__name__)

class OracleBase(ABC):
    """Base class for oracle implementations"""
    
    def __init__(self, name: str, refresh_interval: int = 60):
        """
        Initialize the oracle
        
        Args:
            name: Oracle name
            refresh_interval: Interval in seconds between data fetches
        """
        self.name = name
        self.refresh_interval = refresh_interval
        self.running = False
        self.last_run = None
        self.data = {}
        logger.info(f"Oracle {name} initialized with refresh interval {refresh_interval}s")
    
    @abstractmethod
    async def fetch_data(self) -> Dict[str, Any]:
        """
        Fetch data from external source
        
        Returns:
            Dictionary with fetched data
        """
        pass
    
    @abstractmethod
    async def process_data(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Process the fetched data
        
        Args:
            data: Data fetched from external source
            
        Returns:
            Processed data
        """
        pass
    
    async def run_once(self) -> Dict[str, Any]:
        """
        Run the oracle once
        
        Returns:
            Latest data
        """
        try:
            logger.debug(f"Oracle {self.name} fetching data")
            raw_data = await self.fetch_data()
            
            logger.debug(f"Oracle {self.name} processing data")
            processed_data = await self.process_data(raw_data)
            
            self.data = processed_data
            self.last_run = datetime.utcnow()
            
            return processed_data
        except Exception as e:
            logger.error(f"Oracle {self.name} error: {str(e)}")
            raise
    
    async def run_forever(self):
        """Run the oracle continuously"""
        self.running = True
        
        logger.info(f"Oracle {self.name} starting continuous operation")
        
        while self.running:
            try:
                await self.run_once()
                logger.debug(f"Oracle {self.name} sleeping for {self.refresh_interval}s")
                await asyncio.sleep(self.refresh_interval)
            except Exception as e:
                logger.error(f"Oracle {self.name} error in run_forever: {str(e)}")
                # Sleep before retry to avoid spamming the logs
                await asyncio.sleep(self.refresh_interval)
    
    def stop(self):
        """Stop the oracle"""
        logger.info(f"Oracle {self.name} stopping")
        self.running = False
    
    def get_status(self) -> Dict[str, Any]:
        """
        Get the oracle status
        
        Returns:
            Status information
        """
        return {
            "name": self.name,
            "running": self.running,
            "last_run": self.last_run.isoformat() if self.last_run else None,
            "refresh_interval": self.refresh_interval
        }
    
    def get_latest_data(self) -> Dict[str, Any]:
        """
        Get the latest data
        
        Returns:
            Latest data
        """
        return self.data