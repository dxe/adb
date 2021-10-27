## Build API backend.

FROM golang:latest AS build-api
WORKDIR /src
COPY go.mod go.sum ./
RUN GOFLAGS=-mod=readonly GOPROXY=https://proxy.golang.org go mod download
COPY main.go ./
COPY config config/
COPY google_groups_sync google_groups_sync/
COPY survey_mailer survey_mailer/
COPY international_mailer international_mailer/
COPY mailer mailer/
COPY event_sync event_sync/
COPY form_processor form_processor/
COPY members members/
COPY model model/
COPY discord discord/
COPY mailing_list_signup mailing_list_signup/
RUN CGO_ENABLED=0 go build -o adb


## Build web UI frontend.   

FROM node:lts AS build-ui
WORKDIR /src
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json webpack.config.js ./
COPY frontend frontend/
RUN npm run build


## Assemble composite server container.

FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
RUN addgroup -S adb && adduser -S adb -G adb

WORKDIR /app
COPY run.sh ./
COPY static static/
COPY templates templates/
COPY --from=build-api /src/adb ./
COPY --from=build-ui /src/dist dist/

USER adb

ENTRYPOINT ["./run.sh"]
