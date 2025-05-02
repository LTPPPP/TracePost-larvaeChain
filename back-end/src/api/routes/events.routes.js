const express = require("express");
const router = express.Router();
const { auth, authorize } = require("../middleware/auth.middleware");
const eventsController = require("../controllers/events.controller");

// Protected routes (require authentication)
router.post(
  "/:shipmentId",
  auth,
  authorize(["admin", "manager", "shipper", "warehouse", "customs"]),
  eventsController.recordEvent
);
router.get("/:shipmentId", eventsController.getShipmentEvents);

module.exports = router;
