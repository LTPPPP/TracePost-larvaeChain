const jwt = require("jsonwebtoken");
const bcrypt = require("bcryptjs");
const { v4: uuidv4 } = require("uuid");
const fs = require("fs").promises;
const path = require("path");
const config = require("../../config");
const logger = require("../../utils/logger");

// Simple file-based user storage for demo purposes
const usersFilePath = path.join(config.STORAGE.ROOT_DIR, "users.json");

/**
 * Load users from file
 * @returns {Promise<Array>}
 */
const loadUsers = async () => {
  try {
    await fs.access(usersFilePath);
    const data = await fs.readFile(usersFilePath, "utf8");
    return JSON.parse(data);
  } catch (error) {
    // If file doesn't exist, create an empty users array
    await fs.mkdir(path.dirname(usersFilePath), { recursive: true });
    await fs.writeFile(usersFilePath, JSON.stringify([]));
    return [];
  }
};

/**
 * Save users to file
 * @param {Array} users - Array of user objects
 * @returns {Promise<void>}
 */
const saveUsers = async (users) => {
  await fs.writeFile(usersFilePath, JSON.stringify(users, null, 2));
};

/**
 * Find user by email
 * @param {string} email - User email
 * @returns {Promise<Object|null>}
 */
const findUserByEmail = async (email) => {
  const users = await loadUsers();
  return users.find((user) => user.email === email) || null;
};

/**
 * Find user by ID
 * @param {string} id - User ID
 * @returns {Promise<Object|null>}
 */
const findUserById = async (id) => {
  const users = await loadUsers();
  return users.find((user) => user.id === id) || null;
};

/**
 * Generate JWT token
 * @param {Object} user - User object
 * @returns {string} - JWT token
 */
const generateToken = (user) => {
  return jwt.sign(
    {
      id: user.id,
      email: user.email,
      name: user.name,
      role: user.role,
    },
    config.JWT.SECRET,
    { expiresIn: config.JWT.EXPIRES_IN }
  );
};

/**
 * Auth controller to handle user authentication
 */
exports.register = async (req, res, next) => {
  try {
    const { name, email, password, role = "user" } = req.body;

    // Validate required fields
    if (!name || !email || !password) {
      return res.status(400).json({
        success: false,
        error: "Please provide name, email and password",
      });
    }

    // Check if user already exists
    const existingUser = await findUserByEmail(email);
    if (existingUser) {
      return res.status(400).json({
        success: false,
        error: "User with this email already exists",
      });
    }

    // Validate role
    const allowedRoles = [
      "admin",
      "manager",
      "shipper",
      "warehouse",
      "customs",
      "user",
    ];
    if (!allowedRoles.includes(role)) {
      return res.status(400).json({
        success: false,
        error: `Role must be one of: ${allowedRoles.join(", ")}`,
      });
    }

    // Hash password
    const salt = await bcrypt.genSalt(10);
    const hashedPassword = await bcrypt.hash(password, salt);

    // Create new user
    const newUser = {
      id: uuidv4(),
      name,
      email,
      password: hashedPassword,
      role,
      createdAt: new Date().toISOString(),
    };

    // Save user
    const users = await loadUsers();
    users.push(newUser);
    await saveUsers(users);

    // Generate token
    const token = generateToken(newUser);

    // Return user without password
    const { password: _, ...userWithoutPassword } = newUser;

    res.status(201).json({
      success: true,
      token,
      data: userWithoutPassword,
    });
  } catch (error) {
    logger.error(`Register error: ${error.message}`);
    next(error);
  }
};

exports.login = async (req, res, next) => {
  try {
    const { email, password } = req.body;

    // Validate required fields
    if (!email || !password) {
      return res.status(400).json({
        success: false,
        error: "Please provide email and password",
      });
    }

    // Check if user exists
    const user = await findUserByEmail(email);
    if (!user) {
      return res.status(401).json({
        success: false,
        error: "Invalid credentials",
      });
    }

    // Check if password matches
    const isMatch = await bcrypt.compare(password, user.password);
    if (!isMatch) {
      return res.status(401).json({
        success: false,
        error: "Invalid credentials",
      });
    }

    // Generate token
    const token = generateToken(user);

    // Return user without password
    const { password: _, ...userWithoutPassword } = user;

    res.status(200).json({
      success: true,
      token,
      data: userWithoutPassword,
    });
  } catch (error) {
    logger.error(`Login error: ${error.message}`);
    next(error);
  }
};

exports.getMe = async (req, res, next) => {
  try {
    const user = await findUserById(req.user.id);

    if (!user) {
      return res.status(404).json({
        success: false,
        error: "User not found",
      });
    }

    // Return user without password
    const { password, ...userWithoutPassword } = user;

    res.status(200).json({
      success: true,
      data: userWithoutPassword,
    });
  } catch (error) {
    logger.error(`Get me error: ${error.message}`);
    next(error);
  }
};
