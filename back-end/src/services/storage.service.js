const fs = require("fs").promises;
const path = require("path");
const config = require("../config");
const logger = require("../utils/logger");

class FileStorageService {
  constructor() {
    this.storagePath = config.STORAGE.ROOT_DIR;
    this.shipmentPath = path.join(this.storagePath, "shipments");
    this.eventPath = path.join(this.storagePath, "events");
    this.documentsPath = path.join(this.storagePath, "documents");
    this.documentsContentPath = path.join(this.documentsPath, "content");

    this.initializeStorage();
  }

  async initializeStorage() {
    try {
      await this.ensureDirectory(this.storagePath);
      await this.ensureDirectory(this.shipmentPath);
      await this.ensureDirectory(this.eventPath);
      await this.ensureDirectory(this.documentsPath);
      await this.ensureDirectory(this.documentsContentPath);

      logger.info("Storage system initialized successfully");
    } catch (error) {
      logger.error(`Failed to initialize storage: ${error.message}`);
      throw new Error(`Storage initialization failed: ${error.message}`);
    }
  }

  async ensureDirectory(dirPath) {
    try {
      await fs.access(dirPath);
    } catch (error) {
      await fs.mkdir(dirPath, { recursive: true });
      logger.info(`Created directory: ${dirPath}`);
    }
  }

  async saveShipment(shipment) {
    try {
      if (!shipment.id) {
        throw new Error("Shipment ID is required");
      }

      const filePath = path.join(this.shipmentPath, `${shipment.id}.json`);

      if (!shipment.createdAt) {
        shipment.createdAt = new Date().toISOString();
      }
      if (!shipment.updatedAt) {
        shipment.updatedAt = new Date().toISOString();
      }

      await fs.writeFile(filePath, JSON.stringify(shipment, null, 2));
      logger.info(`Saved shipment to storage: ${shipment.id}`);

      return true;
    } catch (error) {
      logger.error(`Failed to save shipment: ${error.message}`);
      return false;
    }
  }

  async getShipment(shipmentId) {
    try {
      const filePath = path.join(this.shipmentPath, `${shipmentId}.json`);
      const data = await fs.readFile(filePath, "utf8");
      return JSON.parse(data);
    } catch (error) {
      logger.error(`Failed to get shipment ${shipmentId}: ${error.message}`);
      return null;
    }
  }

  async shipmentExists(shipmentId) {
    try {
      const filePath = path.join(this.shipmentPath, `${shipmentId}.json`);
      await fs.access(filePath);
      return true;
    } catch (error) {
      return false;
    }
  }

  async listShipments() {
    try {
      const files = await fs.readdir(this.shipmentPath);
      const shipments = [];

      for (const file of files) {
        if (file.endsWith(".json")) {
          const filePath = path.join(this.shipmentPath, file);
          const data = await fs.readFile(filePath, "utf8");
          shipments.push(JSON.parse(data));
        }
      }

      return shipments;
    } catch (error) {
      logger.error(`Failed to list shipments: ${error.message}`);
      return [];
    }
  }

  async deleteShipment(shipmentId) {
    try {
      const filePath = path.join(this.shipmentPath, `${shipmentId}.json`);
      await fs.unlink(filePath);
      logger.info(`Deleted shipment: ${shipmentId}`);
      return true;
    } catch (error) {
      logger.error(`Failed to delete shipment ${shipmentId}: ${error.message}`);
      return false;
    }
  }

  async saveEvent(event) {
    try {
      if (!event.id || !event.shipmentId) {
        throw new Error("Event ID and Shipment ID are required");
      }

      const filePath = path.join(this.eventPath, `${event.id}.json`);

      if (!event.timestamp) {
        event.timestamp = new Date().toISOString();
      }

      await fs.writeFile(filePath, JSON.stringify(event, null, 2));
      logger.info(
        `Saved event to storage: ${event.id} for shipment ${event.shipmentId}`
      );

      return true;
    } catch (error) {
      logger.error(`Failed to save event: ${error.message}`);
      return false;
    }
  }

  async getShipmentEvents(shipmentId) {
    try {
      const files = await fs.readdir(this.eventPath);
      const events = [];

      for (const file of files) {
        if (file.endsWith(".json")) {
          const filePath = path.join(this.eventPath, file);
          const data = await fs.readFile(filePath, "utf8");
          const event = JSON.parse(data);

          if (event.shipmentId === shipmentId) {
            events.push(event);
          }
        }
      }

      events.sort((a, b) => new Date(a.timestamp) - new Date(b.timestamp));

      return events;
    } catch (error) {
      logger.error(
        `Failed to get events for shipment ${shipmentId}: ${error.message}`
      );
      return [];
    }
  }

  async saveDocument(document, content) {
    try {
      if (!document.id || !document.shipmentId) {
        throw new Error("Document ID and Shipment ID are required");
      }

      const metadataPath = path.join(this.documentsPath, `${document.id}.json`);

      if (!document.createdAt) {
        document.createdAt = new Date().toISOString();
      }

      await fs.writeFile(metadataPath, JSON.stringify(document, null, 2));

      const contentPath = path.join(
        this.documentsContentPath,
        `${document.id}`
      );
      await fs.writeFile(contentPath, content);

      logger.info(
        `Saved document to storage: ${document.id} for shipment ${document.shipmentId}`
      );

      return true;
    } catch (error) {
      logger.error(`Failed to save document: ${error.message}`);
      return false;
    }
  }

  async getDocumentMetadata(documentId) {
    try {
      const filePath = path.join(this.documentsPath, `${documentId}.json`);
      const data = await fs.readFile(filePath, "utf8");
      return JSON.parse(data);
    } catch (error) {
      logger.error(
        `Failed to get document metadata ${documentId}: ${error.message}`
      );
      return null;
    }
  }

  async getDocumentContent(documentId) {
    try {
      const filePath = path.join(this.documentsContentPath, documentId);
      return await fs.readFile(filePath);
    } catch (error) {
      logger.error(
        `Failed to get document content ${documentId}: ${error.message}`
      );
      return null;
    }
  }

  async getShipmentDocuments(shipmentId) {
    try {
      const files = await fs.readdir(this.documentsPath);
      const documents = [];

      for (const file of files) {
        if (file.endsWith(".json")) {
          const filePath = path.join(this.documentsPath, file);
          const data = await fs.readFile(filePath, "utf8");
          const document = JSON.parse(data);

          if (document.shipmentId === shipmentId) {
            documents.push(document);
          }
        }
      }

      return documents;
    } catch (error) {
      logger.error(
        `Failed to get documents for shipment ${shipmentId}: ${error.message}`
      );
      return [];
    }
  }
}

module.exports = new FileStorageService();
