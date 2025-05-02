const shipmentService = require("../../services/shipment.service");
const logger = require("../../utils/logger");

exports.createShipment = async (req, res, next) => {
  try {
    const shipmentData = req.body;

    if (req.user) {
      shipmentData.createdBy = req.user.id;
    }

    const result = await shipmentService.createShipment(shipmentData);

    if (!result.success) {
      return res.status(400).json({
        success: false,
        error: result.error,
      });
    }

    return res.status(201).json({
      success: true,
      data: result.shipment,
    });
  } catch (error) {
    logger.error(`Create shipment error: ${error.message}`);
    next(error);
  }
};

exports.getShipment = async (req, res, next) => {
  try {
    const { id } = req.params;

    const result = await shipmentService.getShipment(id);

    if (!result.success) {
      return res.status(404).json({
        success: false,
        error: result.error,
      });
    }

    return res.status(200).json({
      success: true,
      data: result.shipment,
    });
  } catch (error) {
    logger.error(`Get shipment error: ${error.message}`);
    next(error);
  }
};

exports.getAllShipments = async (req, res, next) => {
  try {
    const result = await shipmentService.getAllShipments();

    return res.status(200).json({
      success: true,
      count: result.shipments.length,
      data: result.shipments,
    });
  } catch (error) {
    logger.error(`Get all shipments error: ${error.message}`);
    next(error);
  }
};

exports.updateShipment = async (req, res, next) => {
  try {
    const { id } = req.params;
    const updateData = req.body;

    if (req.user) {
      updateData.updatedBy = req.user.id;
    }

    const result = await shipmentService.updateShipment(id, updateData);

    if (!result.success) {
      return res.status(404).json({
        success: false,
        error: result.error,
      });
    }

    return res.status(200).json({
      success: true,
      data: result.shipment,
    });
  } catch (error) {
    logger.error(`Update shipment error: ${error.message}`);
    next(error);
  }
};

exports.deleteShipment = async (req, res, next) => {
  try {
    const { id } = req.params;

    const result = await shipmentService.deleteShipment(id);

    if (!result.success) {
      return res.status(404).json({
        success: false,
        error: result.error,
      });
    }

    return res.status(200).json({
      success: true,
      data: {},
    });
  } catch (error) {
    logger.error(`Delete shipment error: ${error.message}`);
    next(error);
  }
};

exports.getShipmentEvents = async (req, res, next) => {
  try {
    const { id } = req.params;

    const result = await shipmentService.getShipmentEvents(id);

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
