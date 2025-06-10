# Stage 1: Build the Next.js application
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files and install dependencies
COPY frontend/package.json frontend/package-lock.json /app/
RUN npm ci

# Copy the rest of the application code
COPY frontend/ /app/

# Build the Next.js application
RUN npm run build

# Stage 2: Create production image
FROM node:22-alpine AS runner

# Set working directory
WORKDIR /app

# Create non-root user for security
RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

# Set environment to production
ENV NODE_ENV=production

# Copy necessary files from builder stage
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

# Set user for running the application
USER nextjs

# Expose frontend port
EXPOSE 3000

# Set host and port environment variables
ENV PORT=3000
ENV HOST=0.0.0.0

# Start the application
CMD ["node", "server.js"]