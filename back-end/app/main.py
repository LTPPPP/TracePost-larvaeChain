# main.py
from fastapi import FastAPI, Depends, Request, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from fastapi.openapi.docs import get_swagger_ui_html
from fastapi.openapi.utils import get_openapi
import time
import uuid
import sys
import os

# Add the parent directory to sys.path if running directly
if __name__ == "__main__":
    sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

# Now import app modules
from app.api.routes import auth, shipments, events, tracing, alerts, admin
from app.core.exceptions import BlockchainLogisticsException
from app.utils.logger import setup_logging, get_logger
from app.config import settings
from app.db.database import init_db
from app.api.middleware import AuditLogMiddleware

# Initialize logging
logger = get_logger(__name__)

# Create FastAPI app
app = FastAPI(
    title="Blockchain Logistics Traceability",
    description="API for tracing logistics operations using blockchain technology",
    version="1.0.0",
    docs_url=None,  # Disable default docs - we'll use custom
    redoc_url=None,  # Disable default redoc
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Add audit log middleware
app.add_middleware(AuditLogMiddleware)

# Add request ID middleware
@app.middleware("http")
async def add_request_id(request: Request, call_next):
    request_id = str(uuid.uuid4())
    request.state.request_id = request_id
    
    # Add process time header
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    
    # Add custom headers
    response.headers["X-Request-ID"] = request_id
    response.headers["X-Process-Time"] = str(process_time)
    
    return response

# Exception handler
@app.exception_handler(BlockchainLogisticsException)
async def blockchain_logistics_exception_handler(request: Request, exc: BlockchainLogisticsException):
    logger.error(f"Error: {exc.detail}", extra={"request_id": getattr(request.state, "request_id", "unknown")})
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content={"detail": exc.detail},
    )

# Include routers
app.include_router(auth.router, prefix="/api/v1", tags=["Authentication"])
app.include_router(shipments.router, prefix="/api/v1", tags=["Shipments"])
app.include_router(events.router, prefix="/api/v1", tags=["Events"])
app.include_router(tracing.router, prefix="/api/v1", tags=["Tracing"])
app.include_router(alerts.router, prefix="/api/v1", tags=["Alerts"])
app.include_router(admin.router, prefix="/api/v1", tags=["Admin"])

# Custom Swagger UI
@app.get("/docs", include_in_schema=False)
async def custom_swagger_ui_html():
    return get_swagger_ui_html(
        openapi_url=app.openapi_url,
        title=f"{app.title} - Swagger UI",
        oauth2_redirect_url=app.swagger_ui_oauth2_redirect_url,
        swagger_js_url="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.18.3/swagger-ui-bundle.js",
        swagger_css_url="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/4.18.3/swagger-ui.css",
        swagger_favicon_url="/favicon.ico",
    )

# Custom OpenAPI schema
def custom_openapi():
    if app.openapi_schema:
        return app.openapi_schema

    openapi_schema = get_openapi(
        title=app.title,
        version=app.version,
        description=app.description,
        routes=app.routes,
    )
    
    openapi_schema["openapi"] = "3.0.2"

    app.openapi_schema = openapi_schema
    return app.openapi_schema

app.openapi = custom_openapi

@app.on_event("startup")
async def startup_event():
    logger.info("Starting Blockchain Logistics Traceability API")
    await init_db()

@app.on_event("shutdown")
async def shutdown_event():
    logger.info("Shutting down Blockchain Logistics Traceability API")

@app.get("/", tags=["Root"])
async def root():
    return {
        "message": "Blockchain Logistics Traceability API",
        "version": app.version,
        "docs": "/docs"
    }

@app.get("/health", tags=["Health"])
async def health_check():
    return {"status": "healthy"}

if __name__ == "__main__":
    import uvicorn
    setup_logging()
    # Run directly with an adjusted path for local development
    uvicorn.run("main:app", host="0.0.0.0", port=7070, reload=settings.DEBUG)