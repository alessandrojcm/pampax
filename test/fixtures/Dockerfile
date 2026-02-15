# Dockerfile for PAMPAX Stage 0.2 Fixture Generation
# Uses Node 22 LTS with proper build tools for native dependencies

FROM node:22-bullseye

# Install build dependencies for native modules
RUN apt-get update && apt-get install -y \
    python3 \
    make \
    g++ \
    git \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /pampax

# Copy package file
COPY package.json ./

# Install pnpm
RUN npm install -g pnpm@8.15.0

# Install dependencies (without frozen lockfile since we don't have one)
RUN pnpm install

# Copy source code
COPY . .

# Set environment variables for reproducible builds
ENV PAMPAX_PROVIDER=transformers
ENV NODE_ENV=production

# Default command
CMD ["/bin/bash"]
