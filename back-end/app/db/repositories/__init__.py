from app.db.repositories.base import BaseRepository
from app.db.repositories.user import (
    UserRepository, 
    OrganizationRepository
)
from app.db.repositories.shipment import (
    ShipmentRepository,
    ShipmentItemRepository,
    DocumentRepository
)
from app.db.repositories.event import (
    EventRepository,
    ShipmentAlertRepository,
    AuditLogRepository
)
from app.db.repositories.alert import (
    AlertRepository,
    AlertRuleRepository,
    AlertSubscriptionRepository
)

# Initialize singleton repository instances
user_repository = UserRepository()
organization_repository = OrganizationRepository()

shipment_repository = ShipmentRepository()
shipment_item_repository = ShipmentItemRepository()
document_repository = DocumentRepository()

event_repository = EventRepository()
shipment_alert_repository = ShipmentAlertRepository()
audit_log_repository = AuditLogRepository()

alert_repository = AlertRepository()
alert_rule_repository = AlertRuleRepository()
alert_subscription_repository = AlertSubscriptionRepository()

__all__ = [
    "BaseRepository",
    # User related
    "user_repository",
    "organization_repository",
    # Shipment related
    "shipment_repository",
    "shipment_item_repository", 
    "document_repository",
    # Event related
    "event_repository",
    "shipment_alert_repository",
    "audit_log_repository",
    # Alert related
    "alert_repository",
    "alert_rule_repository",
    "alert_subscription_repository"
]