const shipmentService = require("../../services/shipment.service");
const logger = require("../../utils/logger");

exports.verifyShipment = async (req, res, next) => {
  try {
    const { shipmentId } = req.params;

    const shipmentResult = await shipmentService.getShipment(shipmentId);

    if (!shipmentResult.success) {
      return res.status(404).json({
        success: false,
        error: shipmentResult.error,
      });
    }

    const blockchainResult = await shipmentService.verifyShipmentOnBlockchain(
      shipmentId
    );

    return res.status(200).json({
      success: true,
      data: {
        shipment: shipmentResult.shipment,
        blockchain: {
          verified: blockchainResult.verified,
          data: blockchainResult.blockchainData,
        },
      },
    });
  } catch (error) {
    logger.error(`Verify shipment error: ${error.message}`);
    next(error);
  }
};

exports.getBlockchainEvents = async (req, res, next) => {
  try {
    const { shipmentId } = req.params;

    const shipmentResult = await shipmentService.getShipment(shipmentId);

    if (!shipmentResult.success) {
      return res.status(404).json({
        success: false,
        error: shipmentResult.error,
      });
    }

    const eventsResult = await shipmentService.getBlockchainEvents(shipmentId);

    if (!eventsResult.success) {
      return res.status(400).json({
        success: false,
        error: eventsResult.error,
      });
    }

    return res.status(200).json({
      success: true,
      count: eventsResult.events.length,
      data: eventsResult.events,
    });
  } catch (error) {
    logger.error(`Get blockchain events error: ${error.message}`);
    next(error);
  }
};
