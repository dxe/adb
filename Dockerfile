## Build API backend.
# Keep in sync with server/src/go.mod.
FROM golang:1.26.0 AS build-api
WORKDIR /workspace/server/src
COPY go.work /workspace/
COPY go.work.sum /workspace/
COPY cli/ /workspace/cli/
COPY jobs/ /workspace/jobs/
COPY server/src ./
COPY pkg/ /workspace/pkg/
RUN GOFLAGS=-mod=readonly GOPROXY=https://proxy.golang.org go mod download
RUN CGO_ENABLED=0 go build -o /adb-server

WORKDIR /workspace/cli
RUN CGO_ENABLED=0 go build -o /adb

## Assemble composite server container.
FROM alpine:latest
ENV ADB_IN_DOCKER=true
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S adb && adduser -S adb -G adb
WORKDIR /app
COPY server/run.sh ./
COPY server/templates templates/
COPY server/static static/
# Ship the CLI in the production image so adb is available for debugging and
# one-off operational tasks inside the running container.
COPY --from=build-api /adb-server ./
COPY --from=build-api /adb ./adb
RUN ln -s /app/adb /usr/local/bin/adb
USER adb
ENTRYPOINT ["./run.sh"]
