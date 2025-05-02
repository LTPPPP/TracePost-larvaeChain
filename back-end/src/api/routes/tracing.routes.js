const express = require("express");
const router = express.Router();
const tracingController = require("../controllers/tracing.controller");

// Public verification routes
router.get("/verify/:shipmentId", tracingController.verifyShipment);
router.get("/events/:shipmentId", tracingController.getBlockchainEvents);

module.exports = router;
