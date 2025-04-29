# __init__.py

# Import all schemas for easy access
from app.schemas.alert import (
    Alert, AlertCreate, AlertUpdate,
    AlertRule, AlertRuleCreate, AlertRuleUpdate,
    AlertSubscription, AlertSubscriptionCreate, AlertSubscriptionUpdate,
    AlertSeverityEnum, AlertTypeEnum, AlertStatusEnum,
    AlertCount, AlertNotification
)

from app.schemas.auth import (
    User, UserCreate, UserUpdate, UserInDB,
    Token, TokenPayload, LoginRequest, RefreshTokenRequest,
    Organization, OrganizationCreate, OrganizationUpdate,
    APIKeyResponse, APIKeyCreate
)

from app.schemas.event import (
    Event, EventCreate, EventUpdate,
    Timeline, TimelineEvent,
    EventTypeEnum, EventSourceEnum,
    EventSummary, ShipmentEventStats
)

from app.schemas.shipment import (
    Shipment, ShipmentCreate, ShipmentUpdate,
    ShipmentItem, ShipmentItemCreate, ShipmentItemUpdate,
    Document, DocumentCreate, DocumentUpdate,
    ShipmentStatusEnum, VerificationStatusEnum,
    ShipmentWithItems, ShipmentWithEvents, ShipmentWithItemsAndEvents,
    ShipmentSummary
)

# For backwards compatibility
from app.schemas.alert import Alert as AlertSchema
from app.schemas.auth import User as UserSchema
from app.schemas.event import Event as EventSchema
from app.schemas.shipment import Shipment as ShipmentSchema