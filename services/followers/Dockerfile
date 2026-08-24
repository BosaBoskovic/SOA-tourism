FROM node:20-alpine

WORKDIR /app

COPY package*.json ./
RUN npm install

COPY src ./src

EXPOSE 8084

# Docker healthcheck oslanja se na /health endpoint (vidi src/observability.js)
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=5 \
    CMD wget -q -O- http://localhost:8084/health || exit 1

CMD ["npm", "start"]
