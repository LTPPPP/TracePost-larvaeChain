FROM node:16-alpine

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Copy source code
COPY . .

# Create storage directory
RUN mkdir -p storage/shipments storage/events storage/documents/content storage/ipfs_cache

# Expose API port
EXPOSE 7070

# Set environment variables
ENV NODE_ENV=production
ENV PORT=7070

# Start the application
CMD ["node", "src/index.js"]