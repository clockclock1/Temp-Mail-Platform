ARG TARGETPLATFORM

FROM --platform=$TARGETPLATFORM node:22-alpine AS frontend-build
WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM --platform=$TARGETPLATFORM golang:1.23-bookworm AS backend-build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tempmail ./cmd/server

FROM --platform=$TARGETPLATFORM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=backend-build /out/tempmail /app/tempmail
COPY --from=frontend-build /frontend/dist /app/web
COPY backend/config/config.example.yaml /app/config/config.yaml

VOLUME ["/app/data", "/app/config"]
EXPOSE 8080 2525
ENTRYPOINT ["/app/tempmail", "-config", "/app/config/config.yaml"]