const shipmentService = require("../../services/shipment.service");
const logger = require("../../utils/logger");

/**
 * Events controller to handle API requests for logistics events
 */
exports.recordEvent = async (req, res, next) => {
  try {
    const { shipmentId } = req.params;
    const eventData = req.body;

    // Add user ID to the event data if authenticated
    if (req.user) {
      eventData.recordedBy = req.user.id;
      eventData.recordedByRole = req.user.role;
    }

    const result = await shipmentService.recordEvent(shipmentId, eventData);

    if (!result.success) {
      return res.status(400).json({
        success: false,
        error: result.error,
      });
    }

    return res.status(201).json({
      success: true,
      data: result.event,
    });
  } catch (error) {
    logger.error(`Record event error: ${error.message}`);
    next(error);
  }
};

exports.getShipmentEvents = async (req, res, next) => {
  try {
    const { shipmentId } = req.params;

    const result = await shipmentService.getShipmentEvents(shipmentId);

    if (!result.success) {
      return res.status(404).json({
        success: false,
        error: result.error,
      });
    }

    return res.status(200).json({
      success: true,
      count: result.events.length,
      data: result.events,
    });
  } catch (error) {
    logger.error(`Get shipment events error: ${error.message}`);
    next(error);
  }
};
