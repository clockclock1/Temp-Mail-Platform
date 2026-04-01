ARG TARGETPLATFORM
FROM --platform=$TARGETPLATFORM golang:1.23-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tempmail ./cmd/server

FROM --platform=$TARGETPLATFORM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/tempmail /app/tempmail

ENV HTTP_ADDR=:8080
ENV SMTP_ADDR=:2525
ENV DB_PATH=/app/data/tempmail.db
ENV DATA_DIR=/app/data/messages

VOLUME ["/app/data"]
EXPOSE 8080 2525
ENTRYPOINT ["/app/tempmail"]