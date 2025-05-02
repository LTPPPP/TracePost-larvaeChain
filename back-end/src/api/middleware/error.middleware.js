const logger = require("../../utils/logger");
const config = require("../../config");

exports.errorHandler = (err, req, res, next) => {
  logger.error(`${err.name}: ${err.message}`, {
    method: req.method,
    url: req.originalUrl,
    stack: err.stack,
  });

  const statusCode = err.statusCode || 500;
  const message = err.message || "Internal server error";

  if (config.NODE_ENV === "production") {
    res.status(statusCode).json({
      success: false,
      error: message,
    });
  } else {
    res.status(statusCode).json({
      success: false,
      error: message,
      stack: err.stack,
    });
  }
};
