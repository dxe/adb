## Build API backend.

FROM golang:latest AS build-api
WORKDIR /src
COPY go.mod go.sum ./
RUN GOFLAGS=-mod=readonly GOPROXY=https://proxy.golang.org go mod download
COPY main.go ./
COPY config config/
COPY mailinglist_sync mailinglist_sync/
COPY survey_mailer survey_mailer/
COPY model model/
RUN CGO_ENABLED=0 go build -o adb


## Build web UI frontend.   

FROM node:latest AS build-ui
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json webpack.config.js ./
COPY frontend frontend/
RUN npm run build


## Assemble composite server container.

FROM alpine:latest
RUN apk add --no-cache tzdata ca-certificates
RUN echo "US/Pacific-New" > /etc/timezone
RUN dpkg-reconfigure -f noninteractive tzdata
RUN addgroup -S adb && adduser -S adb -G adb
USER adb

WORKDIR /app
COPY run.sh ./
COPY static static/
COPY templates templates/
COPY --from=build-api /src/adb ./
COPY --from=build-ui /src/dist dist/

ENTRYPOINT ["./run.sh"]
