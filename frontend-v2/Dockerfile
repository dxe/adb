# syntax=docker/dockerfile:1.7-labs
# Use 1.7-labs to use COPY with --exclude flag.

## Build web UI frontend v2.   

FROM node:20-alpine

WORKDIR /src
COPY --exclude=node_modules --exclude=.next . .

RUN npm install -g pnpm
RUN apk add --no-cache libc6-compat
RUN apk update
ENV PNPM_HOME=/app/.pnpm
ENV PATH=$PNPM_HOME:$PATH

RUN pnpm i --frozen-lockfile
RUN pnpm build

EXPOSE 3000

CMD ["pnpm", "start", "-H", "0.0.0.0"]
