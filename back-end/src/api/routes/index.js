const express = require("express");
const router = express.Router();

// Import route files
const shipmentsRoutes = require("./shipments.routes");
const eventsRoutes = require("./events.routes");
const tracingRoutes = require("./tracing.routes");
const authRoutes = require("./auth.routes");

// Mount routes
router.use("/auth", authRoutes);
router.use("/shipments", shipmentsRoutes);
router.use("/events", eventsRoutes);
router.use("/tracing", tracingRoutes);

// Health check route
router.get("/health", (req, res) => {
  res.status(200).json({
    status: "success",
    message: "API is running",
    timestamp: new Date().toISOString(),
  });
});

module.exports = router;
