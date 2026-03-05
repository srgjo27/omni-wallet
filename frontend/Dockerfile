# ============================================================
# Multi-stage Dockerfile for OmniWallet Frontend
# Stage 1 — deps:     install node_modules (cached separately)
# Stage 2 — builder:  build Next.js app
# Stage 3 — runner:   minimal production image
# ============================================================

# ---- 1. Dependencies ----------------------------------------
FROM node:20-alpine AS deps
WORKDIR /app

# Install OS deps needed by some native modules
RUN apk add --no-cache libc6-compat

COPY package.json package-lock.json* ./
RUN npm ci


# ---- 2. Builder ---------------------------------------------
FROM node:20-alpine AS builder
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Ensure public/ exists even if the directory was not committed (safety net)
RUN mkdir -p /app/public

# Build-time env variable (can be overridden at deploy time)
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

RUN npm run build


# ---- 3. Runner (production) ---------------------------------
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

# Create non-root user
RUN addgroup --system --gid 1001 nodejs \
 && adduser  --system --uid 1001 nextjs

# Copy only the necessary files from builder
COPY --from=builder /app/public         ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static     ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
