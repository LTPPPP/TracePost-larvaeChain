# middleware.py
from typing import Callable
from fastapi import Request, Response
import json
import time
from uuid import UUID

from app.db.database import get_db as get_session  # Import with alias
from app.services import auth as auth_service
from app.utils.logger import get_logger

logger = get_logger(__name__)

class AuditLogMiddleware:
    """
    Middleware for logging all API requests and responses for audit purposes
    """
    
    async def __call__(self, request: Request, call_next: Callable) -> Response:
        # Start timer
        start_time = time.time()
        
        # Get request info
        path = request.url.path
        method = request.method
        client_ip = request.client.host if request.client else "unknown"
        request_id = getattr(request.state, "request_id", "unknown")
        
        # Get user info if available
        user_id = "unknown"
        org_id = "unknown"
        
        # Process request
        response = await call_next(request)
        
        # Calculate processing time
        process_time = time.time() - start_time
        
        # Only log API requests
        if path.startswith("/api/"):
            # Log the request/response
            log_data = {
                "request_id": request_id,
                "method": method,
                "path": path,
                "client_ip": client_ip,
                "user_id": user_id,
                "organization_id": org_id,
                "status_code": response.status_code,
                "process_time_ms": round(process_time * 1000, 2)
            }
            
            # Different log levels based on status code
            if response.status_code >= 500:
                logger.error(f"API Request: {json.dumps(log_data)}")
            elif response.status_code >= 400:
                logger.warning(f"API Request: {json.dumps(log_data)}")
            else:
                logger.info(f"API Request: {json.dumps(log_data)}")
            
            # Store audit log in database for important operations
            if method != "GET" and path not in ["/api/v1/auth/login", "/api/v1/health"]:
                try:
                    async with get_session() as db:
                        # Only log if we have a user ID
                        if user_id != "unknown":
                            await auth_service.create_audit_log(
                                db=db,
                                user_id=UUID(user_id) if user_id != "unknown" else None,
                                action=f"{method} {path}",
                                details=json.dumps(log_data),
                                ip_address=client_ip
                            )
                except Exception as e:
                    logger.error(f"Failed to create audit log: {str(e)}")
        
        return response