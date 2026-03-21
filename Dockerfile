## Build API backend.
# Keep in sync with /workspace/server/src/go.mod.
FROM golang:1.25.0 AS build-api
WORKDIR /workspace/server/src
COPY go.work /workspace/
COPY go.work.sum /workspace/
COPY cli/ /workspace/cli/
COPY server/src ./
COPY pkg/ /workspace/pkg/
RUN GOFLAGS=-mod=readonly GOPROXY=https://proxy.golang.org go mod download
RUN CGO_ENABLED=0 go build -o /adb-server

WORKDIR /workspace/cli
RUN CGO_ENABLED=0 go build -o /adb

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
COPY server/templates templates/
COPY frontend/static static/
# Ship the CLI in the production image so adb is available for debugging and
# one-off operational tasks inside the running container.
COPY --from=build-api /adb-server ./
COPY --from=build-api /adb ./adb
COPY --from=build-ui /src/dist dist/
RUN ln -s /app/adb /usr/local/bin/adb
USER adb
ENTRYPOINT ["./run.sh"]
