# Build stage
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache libc6-compat

# Install Bun globally and Expo CLI
RUN npm install -g bun @expo/cli

# Copy package files
COPY package.json bun.lock* ./

# Install dependencies using Bun
RUN bun install --frozen-lockfile

# Copy source code
COPY . .

# Export for web
RUN bunx expo export -p web

# Production stage
FROM node:20-alpine AS runner

# Set working directory
WORKDIR /app

# Set environment to production
ENV NODE_ENV production

# Install serve to run the static files
RUN npm install -g serve

# Create a non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 expo

# Copy built web app from builder stage
COPY --from=builder /app/dist ./dist

# Change ownership to expo user
RUN chown -R expo:nodejs /app

# Switch to expo user
USER expo

# Expose port 19006
EXPOSE 19006

# Start the application
CMD ["serve", "-s", "dist", "-l", "19006"] 