const express = require("express");
const router = express.Router();
const { auth, authorize } = require("../middleware/auth.middleware");
const shipmentsController = require("../controllers/shipments.controller");

// Public routes
router.get("/:id", shipmentsController.getShipment);
router.get("/", shipmentsController.getAllShipments);

// Protected routes (require authentication)
router.post(
  "/",
  auth,
  authorize(["admin", "manager", "shipper"]),
  shipmentsController.createShipment
);
router.put(
  "/:id",
  auth,
  authorize(["admin", "manager", "shipper"]),
  shipmentsController.updateShipment
);
router.delete(
  "/:id",
  auth,
  authorize(["admin", "manager"]),
  shipmentsController.deleteShipment
);

// Shipment events
router.get("/:id/events", shipmentsController.getShipmentEvents);

module.exports = router;
