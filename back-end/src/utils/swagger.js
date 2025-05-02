const swaggerJSDoc = require("swagger-jsdoc");
const swaggerUi = require("swagger-ui-express");
const config = require("../config");

// Swagger definition
const swaggerDefinition = {
  openapi: "3.0.0",
  info: {
    title: "Blockchain Logistics Traceability API",
    version: "1.0.0",
    description:
      "API documentation for the Blockchain Logistics Traceability System",
    license: {
      name: "MIT",
      url: "https://opensource.org/licenses/MIT",
    },
    contact: {
      name: "API Support",
      email: "support@example.com",
    },
  },
  servers: [
    {
      url: `http://localhost:${config.PORT}`,
      description: "Development server",
    },
    {
      url: "https://api.example.com",
      description: "Production server",
    },
  ],
  components: {
    securitySchemes: {
      bearerAuth: {
        type: "http",
        scheme: "bearer",
        bearerFormat: "JWT",
      },
    },
    schemas: {
      Error: {
        type: "object",
        properties: {
          success: {
            type: "boolean",
            example: false,
          },
          error: {
            type: "string",
            example: "Error message",
          },
        },
      },
      User: {
        type: "object",
        properties: {
          id: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
          name: {
            type: "string",
            example: "John Doe",
          },
          email: {
            type: "string",
            example: "john@example.com",
          },
          role: {
            type: "string",
            enum: [
              "admin",
              "manager",
              "shipper",
              "warehouse",
              "customs",
              "user",
            ],
            example: "manager",
          },
          createdAt: {
            type: "string",
            format: "date-time",
            example: "2023-01-01T00:00:00.000Z",
          },
        },
      },
      Shipment: {
        type: "object",
        properties: {
          id: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
          sender: {
            type: "string",
            example: "Company A",
          },
          receiver: {
            type: "string",
            example: "Company B",
          },
          origin: {
            type: "string",
            example: "Shanghai, China",
          },
          destination: {
            type: "string",
            example: "Rotterdam, Netherlands",
          },
          product: {
            type: "string",
            example: "Electronics",
          },
          quantity: {
            type: "number",
            example: 100,
          },
          unit: {
            type: "string",
            example: "kg",
          },
          status: {
            type: "string",
            enum: [
              "CREATED",
              "IN_TRANSIT",
              "DELIVERED",
              "CUSTOMS_CLEARED",
              "DAMAGED",
              "IN_WAREHOUSE",
              "RETURNED",
            ],
            example: "IN_TRANSIT",
          },
          createdAt: {
            type: "string",
            format: "date-time",
            example: "2023-01-01T00:00:00.000Z",
          },
          updatedAt: {
            type: "string",
            format: "date-time",
            example: "2023-01-02T00:00:00.000Z",
          },
          blockchainVerified: {
            type: "boolean",
            example: true,
          },
          blockchainTxHash: {
            type: "string",
            example:
              "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
          },
          createdBy: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
        },
      },
      Event: {
        type: "object",
        properties: {
          id: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
          shipmentId: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
          type: {
            type: "string",
            enum: [
              "PICKUP",
              "DELIVERY",
              "CUSTOMS_CLEARANCE",
              "WAREHOUSE_IN",
              "WAREHOUSE_OUT",
              "DAMAGED",
              "RETURNED",
            ],
            example: "PICKUP",
          },
          location: {
            type: "string",
            example: "Shanghai Port",
          },
          timestamp: {
            type: "string",
            format: "date-time",
            example: "2023-01-01T00:00:00.000Z",
          },
          data: {
            type: "object",
            example: {
              temperature: "25C",
              humidity: "60%",
              handler: "John Smith",
            },
          },
          recordedBy: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
          recordedByRole: {
            type: "string",
            example: "shipper",
          },
          blockchainTxHash: {
            type: "string",
            example:
              "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
          },
          blockchainVerified: {
            type: "boolean",
            example: true,
          },
        },
      },
      Document: {
        type: "object",
        properties: {
          id: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
          shipmentId: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
          name: {
            type: "string",
            example: "Bill of Lading",
          },
          type: {
            type: "string",
            example: "application/pdf",
          },
          hash: {
            type: "string",
            example:
              "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
          },
          createdAt: {
            type: "string",
            format: "date-time",
            example: "2023-01-01T00:00:00.000Z",
          },
          createdBy: {
            type: "string",
            example: "550e8400-e29b-41d4-a716-446655440000",
          },
        },
      },
      BlockchainVerification: {
        type: "object",
        properties: {
          verified: {
            type: "boolean",
            example: true,
          },
          data: {
            type: "object",
            properties: {
              exists: {
                type: "boolean",
                example: true,
              },
              metadata: {
                type: "object",
                example: {
                  id: "550e8400-e29b-41d4-a716-446655440000",
                  origin: "Shanghai, China",
                  destination: "Rotterdam, Netherlands",
                  product: "Electronics",
                  timestamp: "2023-01-01T00:00:00.000Z",
                },
              },
              registeredBy: {
                type: "string",
                example: "0x1234567890abcdef1234567890abcdef12345678",
              },
              timestamp: {
                type: "number",
                example: 1672531200,
              },
            },
          },
        },
      },
      BlockchainEvent: {
        type: "object",
        properties: {
          type: {
            type: "string",
            example: "PICKUP",
          },
          metadata: {
            type: "object",
            example: {
              id: "550e8400-e29b-41d4-a716-446655440000",
              type: "PICKUP",
              location: "Shanghai Port",
              timestamp: "2023-01-01T00:00:00.000Z",
              data: {
                temperature: "25C",
                humidity: "60%",
              },
            },
          },
          recordedBy: {
            type: "string",
            example: "0x1234567890abcdef1234567890abcdef12345678",
          },
          timestamp: {
            type: "number",
            example: 1672531200,
          },
        },
      },
    },
    responses: {
      UnauthorizedError: {
        description: "Access token is missing or invalid",
        content: {
          "application/json": {
            schema: {
              $ref: "#/components/schemas/Error",
            },
          },
        },
      },
      ForbiddenError: {
        description: "User does not have permission to access the resource",
        content: {
          "application/json": {
            schema: {
              $ref: "#/components/schemas/Error",
            },
          },
        },
      },
      NotFoundError: {
        description: "Resource not found",
        content: {
          "application/json": {
            schema: {
              $ref: "#/components/schemas/Error",
            },
          },
        },
      },
      ValidationError: {
        description: "Validation error",
        content: {
          "application/json": {
            schema: {
              $ref: "#/components/schemas/Error",
            },
          },
        },
      },
      ServerError: {
        description: "Internal server error",
        content: {
          "application/json": {
            schema: {
              $ref: "#/components/schemas/Error",
            },
          },
        },
      },
    },
  },
  tags: [
    {
      name: "Auth",
      description: "Authentication and user management endpoints",
    },
    {
      name: "Shipments",
      description: "Shipment management endpoints",
    },
    {
      name: "Events",
      description: "Logistics event management endpoints",
    },
    {
      name: "Tracing",
      description: "Blockchain tracing and verification endpoints",
    },
  ],
};

// Options for the swagger docs
const options = {
  swaggerDefinition,
  // Path to the API docs
  apis: [
    "./src/api/routes/*.js",
    "./src/api/controllers/*.js",
    "./src/utils/swagger-jsdoc.js",
  ],
};

// Initialize swagger-jsdoc
const swaggerSpec = swaggerJSDoc(options);

/**
 * Setup Swagger middleware for Express
 * @param {Object} app - Express app
 */
const setupSwagger = (app) => {
  // Serve swagger docs
  app.use(
    "/api-docs",
    swaggerUi.serve,
    swaggerUi.setup(swaggerSpec, {
      explorer: true,
      customCss: ".swagger-ui .topbar { display: none }",
      customSiteTitle: "Blockchain Logistics Traceability API Documentation",
    })
  );

  // Serve swagger spec as JSON
  app.get("/swagger.json", (req, res) => {
    res.setHeader("Content-Type", "application/json");
    res.send(swaggerSpec);
  });

  console.log(
    `Swagger docs available at http://localhost:${config.PORT}/api-docs`
  );
};

module.exports = {
  setupSwagger,
  swaggerSpec,
};
