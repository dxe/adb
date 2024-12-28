## Build API backend.

FROM golang:latest AS build-api
WORKDIR /src
COPY server/src ./
RUN GOFLAGS=-mod=readonly GOPROXY=https://proxy.golang.org go mod download
RUN CGO_ENABLED=0 go build -o adb

## Build web UI frontend.   

FROM node:16 AS build-ui

WORKDIR /src
COPY frontend ./
RUN npm ci --legacy-peer-deps
RUN npm run build

## Build web UI frontend v2.   

FROM node:20 AS build-ui-v2

WORKDIR /src
COPY frontend ./
RUN pnpm i --frozen-lockfile
RUN pnpm build

## Assemble composite server container.

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S adb && adduser -S adb -G adb

WORKDIR /app
COPY server/run.sh ./
COPY server/templates templates/
COPY frontend/static static/
COPY --from=build-api /src/adb ./
COPY --from=build-ui /src/dist dist/
COPY --from=build-ui-v2 /src/out js/

USER adb

ENTRYPOINT ["./run.sh"]
