# __init__.py

from app.api.routes import auth, shipments, events, tracing, alerts, admin

__all__ = ["auth", "shipments", "events", "tracing", "alerts", "admin"]