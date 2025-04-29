# Cross-chain bridge
from typing import Dict, Any, Optional, List, Tuple
import asyncio
import json
import hashlib
import time
from datetime import datetime, timedelta

from app.oracle.base import OracleBase
from app.blockchain import ethereum_client, substrate_client, vietnamchain_client
from app.config import settings
from app.utils.logger import get_logger
from app.core.exceptions import BlockchainError

logger = get_logger(__name__)

class ChainBridgeOracle(OracleBase):
    """
    Oracle for bridging data between different blockchains.
    
    This oracle enables cross-chain communication by monitoring events on one
    blockchain and relaying them to another.
    """
    
    def __init__(
        self,
        source_chain: str,
        target_chain: str,
        event_types: Optional[List[str]] = None,
        refresh_interval: int = 300,
        confirmation_blocks: int = 5,
        lookback_hours: int = 24
    ):
        """
        Initialize the chain bridge oracle
        
        Args:
            source_chain: Source blockchain network ("ethereum", "substrate", "vietnamchain")
            target_chain: Target blockchain network ("ethereum", "substrate", "vietnamchain")
            event_types: List of event types to relay, or None for all events
            refresh_interval: Interval in seconds between checks
            confirmation_blocks: Number of blocks required for confirmation
            lookback_hours: Hours to look back for events
        """
        name = f"ChainBridge_{source_chain}_to_{target_chain}"
        super().__init__(name=name, refresh_interval=refresh_interval)
        
        self.source_chain = source_chain.lower()
        self.target_chain = target_chain.lower()
        self.event_types = event_types
        self.confirmation_blocks = confirmation_blocks
        self.lookback_hours = lookback_hours
        self.last_processed_block = None
        self.processed_events = set()
        
        # Select source and target clients
        self.source_client = self._get_client(source_chain)
        self.target_client = self._get_client(target_chain)
        
        if not self.source_client or not self.target_client:
            raise ValueError(f"Both source and target blockchain clients must be configured")
        
        logger.info(f"Chain Bridge Oracle initialized: {source_chain} -> {target_chain}")
    
    def _get_client(self, chain_name: str):
        """Get the blockchain client for a given chain"""
        chain_name = chain_name.lower()
        if chain_name == "ethereum":
            return ethereum_client
        elif chain_name == "substrate":
            return substrate_client
        elif chain_name == "vietnamchain":
            return vietnamchain_client
        else:
            logger.error(f"Unsupported blockchain: {chain_name}")
            return None
    
    async def fetch_data(self) -> Dict[str, Any]:
        """
        Fetch events from the source blockchain
        
        Returns:
            Dictionary with events data
        """
        try:
            # Get events from the source chain
            events = await self._get_recent_events()
            
            return {
                "events": events,
                "timestamp": datetime.utcnow().isoformat(),
                "source_chain": self.source_chain,
                "target_chain": self.target_chain
            }
        except Exception as e:
            logger.error(f"Error fetching data from {self.source_chain}: {str(e)}")
            return {"events": [], "error": str(e)}
    
    async def process_data(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Process events and relay them to the target blockchain
        
        Args:
            data: Data with events from source blockchain
            
        Returns:
            Processing results
        """
        events = data.get("events", [])
        results = []
        
        for event in events:
            try:
                # Skip already processed events
                event_id = event.get("event_id")
                if event_id in self.processed_events:
                    continue
                
                # Relay event to target blockchain
                result = await self._relay_event(event)
                results.append(result)
                
                # Mark as processed
                self.processed_events.add(event_id)
                
            except Exception as e:
                logger.error(f"Error processing event {event.get('event_id')}: {str(e)}")
                results.append({
                    "event_id": event.get("event_id"),
                    "status": "error",
                    "error": str(e)
                })
        
        # Clean up processed events older than 24 hours to prevent memory leaks
        if len(self.processed_events) > 10000:
            self.processed_events = set(list(self.processed_events)[-5000:])
        
        return {
            "processed_count": len(results),
            "results": results,
            "timestamp": datetime.utcnow().isoformat()
        }
    
    async def _get_recent_events(self) -> List[Dict[str, Any]]:
        """
        Get recent events from the source blockchain
        
        Returns:
            List of events
        """
        # This is a simplified implementation
        # In a production system, you would query the blockchain APIs for events
        
        # For Ethereum, you would typically:
        # 1. Get the current block number
        # 2. Calculate the starting block (current - some number of blocks)
        # 3. Fetch events from the contract in that block range
        
        if self.source_chain == "ethereum":
            try:
                # Query the Ethereum client for recent events
                # This would be replaced with actual contract event fetching
                current_block = ethereum_client.web3.eth.block_number
                start_block = self.last_processed_block or (current_block - 1000)
                
                # Don't process blocks that haven't reached confirmation threshold
                end_block = current_block - self.confirmation_blocks
                
                if start_block >= end_block:
                    return []
                
                # Track last processed block
                self.last_processed_block = end_block
                
                # In a real implementation, you would fetch actual events from contracts
                # Here we'll fetch events from your EventLog contract
                events = []
                
                if ethereum_client.event_log:
                    # Example implementation using contract events
                    # This code would need to be adapted to your specific contract structure
                    
                    # Get events filter
                    event_filter = ethereum_client.event_log.contract.events.EventLogged.create_filter(
                        fromBlock=start_block,
                        toBlock=end_block
                    )
                    
                    # Get all entries from the filter
                    entries = event_filter.get_all_entries()
                    
                    for entry in entries:
                        # Process each entry
                        event_data = {
                            "event_id": entry.args.eventId,
                            "shipment_id": entry.args.shipmentId,
                            "event_type": entry.args.eventType,
                            "block_number": entry.blockNumber,
                            "transaction_hash": entry.transactionHash.hex(),
                            "source_chain": self.source_chain,
                            "timestamp": datetime.utcnow().isoformat(),
                            "bridge_id": self._generate_bridge_id(entry.args.eventId)
                        }
                        
                        # Filter by event type if needed
                        if not self.event_types or entry.args.eventType in self.event_types:
                            events.append(event_data)
                
                return events
                
            except Exception as e:
                logger.error(f"Error fetching Ethereum events: {str(e)}")
                return []
                
        elif self.source_chain == "substrate":
            # Implementation for Substrate
            # Would query Substrate node for events
            return []
            
        elif self.source_chain == "vietnamchain":
            # Implementation for VietnamChain
            # Would query VietnamChain API for events
            return []
        
        return []
    
    async def _relay_event(self, event: Dict[str, Any]) -> Dict[str, Any]:
        """
        Relay an event from source to target blockchain
        
        Args:
            event: Event data
            
        Returns:
            Relay result
        """
        try:
            # Prepare data for relay
            event_id = event.get("event_id")
            shipment_id = event.get("shipment_id")
            event_type = event.get("event_type")
            bridge_id = event.get("bridge_id") or self._generate_bridge_id(event_id)
            
            # Create metadata with bridge information
            metadata = json.dumps({
                "source_chain": self.source_chain,
                "source_tx_hash": event.get("transaction_hash"),
                "source_block": event.get("block_number"),
                "bridge_id": bridge_id,
                "original_event_id": event_id,
                "bridged_at": datetime.utcnow().isoformat()
            })
            
            # Create a hash of the event data
            data_str = f"{shipment_id}:{event_id}:{event_type}:{bridge_id}"
            data_hash = hashlib.sha256(data_str.encode()).hexdigest()
            
            # Register on target blockchain
            tx_hash = await self.target_client.register_event(
                shipment_id=shipment_id,
                event_id=bridge_id,  # Use bridge_id as the new event_id
                event_type=f"BRIDGED_{event_type}",
                data_hash=data_hash,
                metadata=metadata
            )
            
            logger.info(f"Event {event_id} bridged from {self.source_chain} to {self.target_chain} with tx {tx_hash}")
            
            return {
                "original_event_id": event_id,
                "bridge_id": bridge_id,
                "shipment_id": shipment_id,
                "source_chain": self.source_chain,
                "target_chain": self.target_chain,
                "target_tx_hash": tx_hash,
                "status": "success"
            }
            
        except Exception as e:
            logger.error(f"Error relaying event {event.get('event_id')}: {str(e)}")
            return {
                "original_event_id": event.get("event_id"),
                "shipment_id": event.get("shipment_id"),
                "source_chain": self.source_chain,
                "target_chain": self.target_chain,
                "status": "error",
                "error": str(e)
            }
    
    def _generate_bridge_id(self, original_id: str) -> str:
        """
        Generate a bridge ID for an event
        
        Args:
            original_id: Original event ID
            
        Returns:
            Bridge ID
        """
        timestamp = int(time.time())
        bridge_data = f"{original_id}:{self.source_chain}:{self.target_chain}:{timestamp}"
        bridge_hash = hashlib.sha256(bridge_data.encode()).hexdigest()[:16]
        return f"bridge_{self.source_chain}_{bridge_hash}"
    
    async def verify_bridged_event(
        self, 
        bridge_id: str, 
        original_event_id: str
    ) -> Dict[str, Any]:
        """
        Verify a bridged event exists on both chains
        
        Args:
            bridge_id: Bridge ID
            original_event_id: Original event ID
            
        Returns:
            Verification result
        """
        try:
            # Verify on source chain
            source_result = await self.source_client.verify_event(
                shipment_id="",  # This would need to be populated in a real implementation
                event_id=original_event_id
            )
            
            # Verify on target chain
            target_result = await self.target_client.verify_event(
                shipment_id="",  # This would need to be populated in a real implementation
                event_id=bridge_id
            )
            
            # Check if both verifications succeeded
            source_verified = source_result.get("verified", False)
            target_verified = target_result.get("verified", False)
            
            return {
                "bridge_id": bridge_id,
                "original_event_id": original_event_id,
                "source_chain": self.source_chain,
                "target_chain": self.target_chain,
                "source_verified": source_verified,
                "target_verified": target_verified,
                "verified": source_verified and target_verified,
                "source_details": source_result,
                "target_details": target_result
            }
            
        except Exception as e:
            logger.error(f"Error verifying bridged event {bridge_id}: {str(e)}")
            return {
                "bridge_id": bridge_id,
                "original_event_id": original_event_id,
                "source_chain": self.source_chain,
                "target_chain": self.target_chain,
                "verified": False,
                "error": str(e)
            }


# Factory to create cross-chain bridges
class ChainBridgeFactory:
    """Factory for creating chain bridges"""
    
    @staticmethod
    def create_bridge(
        source_chain: str,
        target_chain: str,
        event_types: Optional[List[str]] = None,
        refresh_interval: int = 300
    ) -> ChainBridgeOracle:
        """
        Create a new chain bridge
        
        Args:
            source_chain: Source blockchain
            target_chain: Target blockchain
            event_types: List of event types to relay
            refresh_interval: Refresh interval in seconds
            
        Returns:
            Chain bridge oracle
        """
        bridge = ChainBridgeOracle(
            source_chain=source_chain,
            target_chain=target_chain,
            event_types=event_types,
            refresh_interval=refresh_interval
        )
        
        return bridge
    
    @staticmethod
    def create_two_way_bridge(
        chain_a: str,
        chain_b: str,
        event_types: Optional[List[str]] = None,
        refresh_interval: int = 300
    ) -> Tuple[ChainBridgeOracle, ChainBridgeOracle]:
        """
        Create a two-way bridge between two chains
        
        Args:
            chain_a: First blockchain
            chain_b: Second blockchain
            event_types: List of event types to relay
            refresh_interval: Refresh interval in seconds
            
        Returns:
            Tuple of (a_to_b_bridge, b_to_a_bridge)
        """
        a_to_b = ChainBridgeFactory.create_bridge(
            source_chain=chain_a,
            target_chain=chain_b,
            event_types=event_types,
            refresh_interval=refresh_interval
        )
        
        b_to_a = ChainBridgeFactory.create_bridge(
            source_chain=chain_b,
            target_chain=chain_a,
            event_types=event_types,
            refresh_interval=refresh_interval
        )
        
        return (a_to_b, b_to_a)


# Bridge manager to handle multiple bridges
class ChainBridgeManager:
    """Manager for multiple chain bridges"""
    
    def __init__(self):
        """Initialize the chain bridge manager"""
        self.bridges = {}
        self.tasks = {}
        logger.info("Chain Bridge Manager initialized")
    
    def add_bridge(self, bridge: ChainBridgeOracle) -> str:
        """
        Add a bridge to the manager
        
        Args:
            bridge: Chain bridge oracle
            
        Returns:
            Bridge ID
        """
        bridge_id = bridge.name
        self.bridges[bridge_id] = bridge
        logger.info(f"Bridge {bridge_id} added to manager")
        return bridge_id
    
    def remove_bridge(self, bridge_id: str) -> bool:
        """
        Remove a bridge from the manager
        
        Args:
            bridge_id: Bridge ID
            
        Returns:
            True if removed, False otherwise
        """
        if bridge_id in self.bridges:
            # Stop the bridge if running
            self.stop_bridge(bridge_id)
            
            # Remove bridge
            del self.bridges[bridge_id]
            logger.info(f"Bridge {bridge_id} removed from manager")
            return True
        
        return False
    
    async def start_bridge(self, bridge_id: str) -> bool:
        """
        Start a bridge
        
        Args:
            bridge_id: Bridge ID
            
        Returns:
            True if started, False otherwise
        """
        if bridge_id in self.bridges and bridge_id not in self.tasks:
            bridge = self.bridges[bridge_id]
            
            # Create task
            task = asyncio.create_task(bridge.run_forever())
            self.tasks[bridge_id] = task
            
            logger.info(f"Bridge {bridge_id} started")
            return True
        
        return False
    
    def stop_bridge(self, bridge_id: str) -> bool:
        """
        Stop a bridge
        
        Args:
            bridge_id: Bridge ID
            
        Returns:
            True if stopped, False otherwise
        """
        if bridge_id in self.bridges and bridge_id in self.tasks:
            bridge = self.bridges[bridge_id]
            task = self.tasks[bridge_id]
            
            # Stop bridge
            bridge.stop()
            
            # Cancel task
            task.cancel()
            del self.tasks[bridge_id]
            
            logger.info(f"Bridge {bridge_id} stopped")
            return True
        
        return False
    
    def is_bridge_running(self, bridge_id: str) -> bool:
        """
        Check if a bridge is running
        
        Args:
            bridge_id: Bridge ID
            
        Returns:
            True if running, False otherwise
        """
        return bridge_id in self.tasks
    
    def get_bridge(self, bridge_id: str) -> Optional[ChainBridgeOracle]:
        """
        Get a bridge by ID
        
        Args:
            bridge_id: Bridge ID
            
        Returns:
            Chain bridge oracle or None
        """
        return self.bridges.get(bridge_id)
    
    def get_all_bridges(self) -> Dict[str, ChainBridgeOracle]:
        """
        Get all bridges
        
        Returns:
            Dictionary of bridge ID to bridge
        """
        return self.bridges
    
    def get_bridge_status(self, bridge_id: str) -> Optional[Dict[str, Any]]:
        """
        Get bridge status
        
        Args:
            bridge_id: Bridge ID
            
        Returns:
            Bridge status or None
        """
        bridge = self.get_bridge(bridge_id)
        if not bridge:
            return None
        
        status = bridge.get_status()
        status["running"] = self.is_bridge_running(bridge_id)
        
        return status
    
    def get_all_statuses(self) -> Dict[str, Dict[str, Any]]:
        """
        Get status of all bridges
        
        Returns:
            Dictionary of bridge ID to status
        """
        return {
            bridge_id: self.get_bridge_status(bridge_id)
            for bridge_id in self.bridges
        }
    
    async def start_all_bridges(self) -> Dict[str, bool]:
        """
        Start all bridges
        
        Returns:
            Dictionary of bridge ID to success status
        """
        results = {}
        for bridge_id in self.bridges:
            results[bridge_id] = await self.start_bridge(bridge_id)
        
        return results
    
    def stop_all_bridges(self) -> Dict[str, bool]:
        """
        Stop all bridges
        
        Returns:
            Dictionary of bridge ID to success status
        """
        results = {}
        for bridge_id in list(self.tasks.keys()):
            results[bridge_id] = self.stop_bridge(bridge_id)
        
        return results


# Create global bridge manager instance
bridge_manager = ChainBridgeManager()