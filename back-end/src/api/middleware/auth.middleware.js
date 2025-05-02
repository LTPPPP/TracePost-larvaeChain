const jwt = require("jsonwebtoken");
const config = require("../../config");
const logger = require("../../utils/logger");

exports.auth = (req, res, next) => {
  try {
    const authHeader = req.headers.authorization;

    if (!authHeader || !authHeader.startsWith("Bearer ")) {
      return res.status(401).json({
        success: false,
        error: "Authentication required. Please provide a valid token.",
      });
    }

    const token = authHeader.split(" ")[1];

    if (!token) {
      return res.status(401).json({
        success: false,
        error: "Authentication required. Please provide a valid token.",
      });
    }

    const decoded = jwt.verify(token, config.JWT.SECRET);

    req.user = decoded;

    next();
  } catch (error) {
    logger.error(`Auth Middleware Error: ${error.message}`);

    if (error.name === "TokenExpiredError") {
      return res.status(401).json({
        success: false,
        error: "Token expired. Please login again.",
      });
    }

    if (error.name === "JsonWebTokenError") {
      return res.status(401).json({
        success: false,
        error: "Invalid token. Please login again.",
      });
    }

    return res.status(401).json({
      success: false,
      error: "Authentication failed. Please login again.",
    });
  }
};

exports.authorize = (roles = []) => {
  return (req, res, next) => {
    if (!req.user) {
      return res.status(401).json({
        success: false,
        error: "Authentication required",
      });
    }

    if (roles.length > 0 && !roles.includes(req.user.role)) {
      return res.status(403).json({
        success: false,
        error: "Not authorized to access this resource",
      });
    }

    next();
  };
};
