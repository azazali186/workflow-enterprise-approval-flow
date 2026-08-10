# syntax=docker/dockerfile:1

# ---------- Builder ----------
FROM node:20-alpine AS builder

WORKDIR /app

# Cache npm dependencies
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# The API base URL is baked into the bundle at build time. Keep the relative
# path ("/api/v1") in production so nginx can proxy to the backend on the same
# origin (no CORS needed); override for CDN-based hosting.
ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

RUN npm run build

# ---------- Runtime ----------
FROM nginx:1.27-alpine

# The official nginx image renders /etc/nginx/templates/*.template with
# envsubst at startup, substituting only defined env vars — nginx's own
# $http_upgrade / $proxy_add_x_forwarded_for variables are left untouched.
COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=builder /app/dist /usr/share/nginx/html

# Override in compose/k8s to point at the backend service.
ENV BACKEND_UPSTREAM=http://approval-flow:8080

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://127.0.0.1/ >/dev/null 2>&1 || exit 1
