from typing import List, Dict, Any, Optional
from uuid import UUID
from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, BackgroundTasks, Query, Path, Body
from sqlalchemy.ext.asyncio import AsyncSession

from app.api.dependencies import get_db, get_current_admin_user
from app.services import tracing as tracing_service
from app.models.user import User
from app.blockchain import ethereum_client, substrate_client, vietnamchain_client
from app.core.exceptions import BlockchainError
from app.schemas.blockchain import ContractDeploymentRequest, BlockchainStatus

router = APIRouter(prefix="/admin", tags=["admin"])

@router.get("/blockchain/status", response_model=Dict[str, Any])
async def get_blockchain_status(
    current_user: User = Depends(get_current_admin_user)
):
    """Get blockchain services status"""
    status = {
        "ethereum": {
            "enabled": ethereum_client is not None,
            "chain_id": ethereum_client.chain_id if ethereum_client else None,
            "network": ethereum_client._get_network_name() if ethereum_client else None,
            "node_url": ethereum_client.node_url if ethereum_client else None,
            "account": ethereum_client.address if ethereum_client else None,
        },
        "substrate": {
            "enabled": substrate_client is not None,
            "network": substrate_client.network_name if substrate_client else None,
            "node_url": substrate_client.node_url if substrate_client else None,
        },
        "vietnamchain": {
            "enabled": vietnamchain_client is not None,
            "node_url": vietnamchain_client.node_url if vietnamchain_client else None,
        },
        "timestamp": datetime.utcnow().isoformat()
    }
    
    return status

@router.post("/blockchain/contracts/deploy", response_model=Dict[str, Any])
async def deploy_contract(
    deployment: ContractDeploymentRequest,
    background_tasks: BackgroundTasks,
    current_user: User = Depends(get_current_admin_user)
):
    """Deploy a contract to the blockchain"""
    
    # Select blockchain client
    client = None
    if deployment.network.lower() == "ethereum":
        client = ethereum_client
    elif deployment.network.lower() == "substrate":
        client = substrate_client
    elif deployment.network.lower() == "vietnamchain":
        client = vietnamchain_client
    else:
        raise HTTPException(status_code=400, detail=f"Unsupported blockchain network: {deployment.network}")
    
    if not client:
        raise HTTPException(status_code=400, detail=f"Blockchain client for {deployment.network} is not configured")
    
    try:
        # For Ethereum, we have a full deployment flow
        if deployment.network.lower() == "ethereum":
            # Compile contract
            contract_interface, bytecode = ethereum_client.compile_solidity_contract(
                contract_name=deployment.contract_name,
                solidity_version=deployment.compiler_version
            )
            
            # Deploy contract (in background if specified)
            if deployment.deploy_in_background:
                background_tasks.add_task(
                    ethereum_client.deploy_contract,
                    contract_name=deployment.contract_name,
                    constructor_args=deployment.constructor_args
                )
                
                return {
                    "status": "pending",
                    "contract_name": deployment.contract_name,
                    "network": deployment.network,
                    "message": "Contract deployment started in background"
                }
            else:
                contract_address, contract = await ethereum_client.deploy_contract(
                    contract_name=deployment.contract_name,
                    constructor_args=deployment.constructor_args
                )
                
                return {
                    "status": "success",
                    "contract_name": deployment.contract_name,
                    "contract_address": contract_address,
                    "network": deployment.network,
                    "message": f"Contract deployed at {contract_address}"
                }
        else:
            # For other networks, we'll just return the ABI
            contract_interface, _ = client.compile_solidity_contract(
                contract_name=deployment.contract_name,
                solidity_version=deployment.compiler_version
            )
            
            return {
                "status": "not_supported",
                "contract_name": deployment.contract_name,
                "network": deployment.network,
                "message": f"Automatic deployment not supported for {deployment.network}",
                "contract_abi": contract_interface
            }
            
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Contract deployment failed: {str(e)}")

@router.get("/blockchain/contracts/{network}", response_model=Dict[str, Any])
async def get_contract_deployments(
    network: str = Path(..., description="Blockchain network"),
    current_user: User = Depends(get_current_admin_user)
):
    """Get deployed contracts for a network"""
    
    # Select blockchain client
    client = None
    if network.lower() == "ethereum":
        client = ethereum_client
    elif network.lower() == "substrate":
        client = substrate_client
    elif network.lower() == "vietnamchain":
        client = vietnamchain_client
    else:
        raise HTTPException(status_code=400, detail=f"Unsupported blockchain network: {network}")
    
    if not client:
        raise HTTPException(status_code=400, detail=f"Blockchain client for {network} is not configured")
    
    try:
        if network.lower() == "ethereum":
            deployments = await ethereum_client.get_contract_deployments()
            return deployments
        else:
            return {"contracts": [], "message": f"Deployment tracking not supported for {network}"}
            
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get contract deployments: {str(e)}")

@router.post("/blockchain/verify/shipment/{shipment_id}", response_model=Dict[str, Any])
async def verify_shipment_on_blockchain(
    shipment_id: UUID,
    network: str = Query("ethereum", description="Blockchain network to use"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin_user)
):
    """Register a shipment on the blockchain for verification"""
    try:
        result = await tracing_service.verify_shipment_on_blockchain(
            db=db,
            shipment_id=shipment_id,
            user=current_user,
            network=network
        )
        return result
        
    except BlockchainError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to verify shipment: {str(e)}")

@router.post("/blockchain/verify/event/{event_id}", response_model=Dict[str, Any])
async def verify_event_on_blockchain(
    event_id: UUID,
    network: str = Query("ethereum", description="Blockchain network to use"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin_user)
):
    """Register an event on the blockchain for verification"""
    try:
        result = await tracing_service.verify_event_on_blockchain(
            db=db,
            event_id=event_id,
            user=current_user,
            network=network
        )
        return result
        
    except BlockchainError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to verify event: {str(e)}")

@router.get("/blockchain/transaction/{tx_hash}", response_model=Dict[str, Any])
async def get_transaction_status(
    tx_hash: str,
    network: str = Query("ethereum", description="Blockchain network"),
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin_user)
):
    """Get status of a blockchain transaction"""
    try:
        result = await tracing_service.verify_blockchain_transaction(
            db=db,
            tx_hash=tx_hash,
            network=network
        )
        return result
        
    except BlockchainError as e:
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Failed to get transaction status: {str(e)}")