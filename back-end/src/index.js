require("dotenv").config();
const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const morgan = require("morgan");
const { errorHandler } = require("./api/middleware/error.middleware");
const routes = require("./api/routes");
const logger = require("./utils/logger");
const { setupSwagger } = require("./utils/swagger");

const app = express();
const PORT = process.env.PORT || 7070;

app.use(helmet());
app.use(cors());
app.use(express.json());
app.use(
  morgan("combined", {
    stream: { write: (message) => logger.info(message.trim()) },
  })
);

setupSwagger(app);

app.use("/api/v1", routes);

app.use(errorHandler);

app.listen(PORT, () => {
  logger.info(`Server running on port ${PORT}`);
  console.log(`Server running on port ${PORT}`);
  console.log(
    `Swagger documentation available at http://localhost:${PORT}/api-docs`
  );
});

process.on("unhandledRejection", (err) => {
  logger.error("Unhandled Promise Rejection:", err);
  console.error("Unhandled Promise Rejection:", err);
});

module.exports = app;
