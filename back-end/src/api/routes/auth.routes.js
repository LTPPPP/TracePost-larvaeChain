const express = require("express");
const router = express.Router();
const authController = require("../controllers/auth.controller");
const { auth } = require("../middleware/auth.middleware");

// Authentication routes
router.post("/register", authController.register);
router.post("/login", authController.login);
router.get("/me", auth, authController.getMe);

module.exports = router;
