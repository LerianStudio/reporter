# Use an official Node.js runtime as the base image
FROM node:22-alpine AS builder

# Set the working directory in the container
WORKDIR /usr/src/app

# Install dependencies first (better layer caching)
COPY components/frontend/package*.json ./

RUN apk update && apk add bash && \
    npm ci --only=production=false && \
    npm cache clean --force

# Copy source code
COPY components/frontend/. .

# Copy .env.example and set env
RUN cp .env.example .env

# Build the application
RUN npm run build

# Stage 2: Runtime - Minimal production image
FROM node:22-alpine

WORKDIR /usr/src/app

# Install only bash for entrypoint script
RUN apk update && apk add --no-cache bash && \
    rm -rf /var/cache/apk/*

# Copy only production dependencies
COPY components/frontend/package*.json ./

RUN npm ci --only=production && \
    npm cache clean --force

# Copy built artifacts from builder
COPY --from=builder /usr/src/app/.next ./.next
COPY --from=builder /usr/src/app/public ./public
COPY --from=builder /usr/src/app/.env .env

# Copy entrypoint script
COPY components/frontend/scripts/entrypoint.sh /scripts/entrypoint.sh
RUN chmod +x /scripts/entrypoint.sh

# Copy other necessary config files
COPY components/frontend/next.config.mjs ./next.config.mjs
COPY components/frontend/intl.config.ts ./intl.config.ts
COPY components/frontend/locales ./locales

# Expose the port
EXPOSE 8083

# Run as non-root user for security
USER node

# Start the application
ENTRYPOINT ["/scripts/entrypoint.sh"]
CMD ["npm", "run", "start"]
