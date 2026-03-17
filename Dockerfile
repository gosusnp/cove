# Copyright (c) 2026 Jimmy Ma
# SPDX-License-Identifier: Elastic-2.0

# Stage 1: build frontend
FROM node:22-alpine AS frontend-build
WORKDIR /workspace
COPY frontend/package*.json ./frontend/
RUN npm --prefix frontend ci
COPY frontend/ ./frontend/
# vite outDir is ../backend/ui → /workspace/backend/ui
RUN npm --prefix frontend run build

# Stage 2: build backend
FROM golang:1.26-alpine AS backend-build
ARG TARGETARCH
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend-build /workspace/backend/ui ./ui
RUN CGO_ENABLED=0 GOARCH=${TARGETARCH} go build -o bin/cove .

# Stage 3: minimal runtime image
FROM gcr.io/distroless/static-debian13
COPY --from=backend-build /app/bin/cove /cove
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/cove"]
