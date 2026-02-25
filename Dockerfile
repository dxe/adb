## Build API backend.
FROM golang:latest AS build-api
WORKDIR /src
COPY server/src ./
RUN GOFLAGS=-mod=readonly GOPROXY=https://proxy.golang.org go mod download
RUN CGO_ENABLED=0 go build -o adb-server

## Build web UI frontend.
# Please keep Node version in sync with Makefile
FROM node:16 AS build-ui
WORKDIR /src
# Copy shared directory at the correct relative path
COPY shared ../shared/
COPY frontend ./
RUN npm ci --legacy-peer-deps
RUN npm run build

## Assemble composite server container.
FROM alpine:latest
ENV ADB_IN_DOCKER=true
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S adb && adduser -S adb -G adb
WORKDIR /app
COPY server/run.sh ./
COPY server/scripts/db-migrations db-migrations/
COPY server/templates templates/
COPY frontend/static static/
COPY --from=build-api /src/adb-server ./
COPY --from=build-ui /src/dist dist/
USER adb
ENTRYPOINT ["./run.sh"]
