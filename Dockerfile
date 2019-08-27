## Build API backend.

FROM golang AS build-api
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY config config/
COPY mailinglist_sync mailinglist_sync/
COPY model model/
RUN CGO_ENABLED=0 go build -o adb


## Build web UI frontend.   

FROM node AS build-ui
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json webpack.config.js ./
COPY frontend frontend/
RUN npm run build


## Assemble composite server container.

FROM alpine
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY static static/
COPY templates templates/
COPY adb-config/client_secrets.json adb-config/client_secrets.json
COPY --from=build-api /src/adb ./
COPY --from=build-ui /src/dist dist/

ENTRYPOINT ["./adb"]
