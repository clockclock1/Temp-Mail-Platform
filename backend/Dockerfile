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

ENV HTTP_ADDR=:8080
ENV SMTP_ADDR=:2525
ENV WEB_DIR=/app/web
ENV DB_PATH=/app/data/tempmail.db
ENV DATA_DIR=/app/data/messages

VOLUME ["/app/data"]
EXPOSE 8080 2525
ENTRYPOINT ["/app/tempmail"]