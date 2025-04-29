from typing import List, Optional, Dict, Any, Tuple
from uuid import UUID
from datetime import datetime, timedelta
import json

from sqlalchemy.ext.asyncio import AsyncSession
from fastapi import Depends, Request, Response

from app.db.repositories import audit_log_repository
from app.models.event import AuditLog
from app.models.user import User
from app.utils.logger import get_logger

logger = get_logger(__name__)

async def log_api_call(
    db: AsyncSession,
    request: Request,
    response: Response,
    start_time: float,
    user: Optional[User] = None,
    action: Optional[str] = None,
    resource_type: Optional[str] = None,
    resource_id: Optional[str] = None,
    metadata: Optional[Dict[str, Any]] = None
) -> AuditLog:
    """
    Log an API call for auditing purposes
    
    Args:
        db: Database session
        request: HTTP request
        response: HTTP response
        start_time: Start time of request processing
        user: Authenticated user if available
        action: Action performed
        resource_type: Resource type
        resource_id: Resource ID
        metadata: Additional metadata
        
    Returns:
        Created audit log entry
    """
    # Calculate response time
    response_time_ms = (datetime.utcnow().timestamp() - start_time) * 1000
    
    # Get request information
    request_method = request.method
    request_path = request.url.path
    request_id = request.headers.get("X-Request-ID")
    request_ip = request.client.host if request.client else None
    request_user_agent = request.headers.get("User-Agent")
    
    # Get user information if available
    user_id = user.id if user else None
    organization_id = user.organization_id if user else None
    
    # Get response status
    response_status = response.status_code
    
    # Create audit log entry
    log_entry = await audit_log_repository.log_api_call(
        db,
        request_method=request_method,
        request_path=request_path,
        request_id=request_id,
        request_ip=request_ip,
        request_user_agent=request_user_agent,
        response_status=response_status,
        response_time_ms=response_time_ms,
        user_id=user_id,
        organization_id=organization_id,
        action=action,
        resource_type=resource_type,
        resource_id=resource_id,
        metadata=metadata
    )
    
    return log_entry

async def search_audit_logs(
    db: AsyncSession,
    user: User,
    user_id: Optional[UUID] = None,
    organization_id: Optional[UUID] = None,
    action: Optional[str] = None,
    resource_type: Optional[str] = None,
    resource_id: Optional[str] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    skip: int = 0,
    limit: int = 100
) -> List[AuditLog]:
    """
    Search audit logs with filtering
    
    Args:
        db: Database session
        user: Current user
        user_id: Filter by user ID
        organization_id: Filter by organization ID
        action: Filter by action
        resource_type: Filter by resource type
        resource_id: Filter by resource ID
        start_date: Filter by created_at start
        end_date: Filter by created_at end
        skip: Number of records to skip
        limit: Maximum number of records to return
        
    Returns:
        List of audit logs
    """
    # Only admins can see all audit logs
    if user.role != "admin":
        # Non-admins can only see logs for their organization
        organization_id = user.organization_id
    
    return await audit_log_repository.search_logs(
        db,
        user_id=user_id,
        organization_id=organization_id,
        action=action,
        resource_type=resource_type,
        resource_id=resource_id,
        start_date=start_date,
        end_date=end_date,
        skip=skip,
        limit=limit
    )

async def get_recent_audit_logs(
    db: AsyncSession,
    user: User,
    hours: int = 24,
    limit: int = 100
) -> List[AuditLog]:
    """
    Get recent audit logs
    
    Args:
        db: Database session
        user: Current user
        hours: Number of hours to look back
        limit: Maximum number of records to return
        
    Returns:
        List of recent audit logs
    """
    # Calculate start date
    start_date = datetime.utcnow() - timedelta(hours=hours)
    
    # Only admins can see all audit logs
    if user.role != "admin":
        # Non-admins can only see logs for their organization
        organization_id = user.organization_id
    else:
        organization_id = None
    
    return await audit_log_repository.search_logs(
        db,
        organization_id=organization_id,
        start_date=start_date,
        limit=limit
    )

async def get_audit_log_stats(
    db: AsyncSession,
    user: User,
    days: int = 30
) -> Dict[str, Any]:
    """
    Get audit log statistics
    
    Args:
        db: Database session
        user: Current user
        days: Number of days to include in statistics
        
    Returns:
        Dictionary with audit log statistics
    """
    # Calculate start date
    start_date = datetime.utcnow() - timedelta(days=days)
    
    # Only admins can see all audit logs
    if user.role != "admin":
        # Non-admins can only see logs for their organization
        organization_id = user.organization_id
    else:
        organization_id = None
    
    # Get logs for the time period
    logs = await audit_log_repository.search_logs(
        db,
        organization_id=organization_id,
        start_date=start_date
    )
    
    # Calculate statistics
    stats = {
        "total_requests": len(logs),
        "by_method": {},
        "by_status": {},
        "by_action": {},
        "by_resource_type": {},
        "avg_response_time_ms": 0
    }
    
    # Calculate detailed stats
    total_response_time = 0
    
    for log in logs:
        # Count by method
        method = log.request_method
        stats["by_method"][method] = stats["by_method"].get(method, 0) + 1
        
        # Count by status
        status = log.response_status
        status_key = f"{status}"
        stats["by_status"][status_key] = stats["by_status"].get(status_key, 0) + 1
        
        # Count by action
        if log.action:
            stats["by_action"][log.action] = stats["by_action"].get(log.action, 0) + 1
        
        # Count by resource type
        if log.resource_type:
            stats["by_resource_type"][log.resource_type] = stats["by_resource_type"].get(log.resource_type, 0) + 1
        
        # Sum response time
        if log.response_time_ms:
            total_response_time += log.response_time_ms
    
    # Calculate average response time
    if logs:
        stats["avg_response_time_ms"] = total_response_time / len(logs)
    
    return stats