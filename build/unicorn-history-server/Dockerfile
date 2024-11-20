ARG NODE_VERSION=22
ARG ALPINE_VERSION=3.20

FROM node:${NODE_VERSION} AS builder
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /build/src
COPY web ./

RUN npm install -g @angular/cli@18
RUN npm install -g pnpm --force

RUN pnpm install
RUN pnpm run build

FROM alpine:${ALPINE_VERSION}

COPY --from=builder /build/assets /app/assets
COPY bin/app/unicorn-history-server /app/unicorn-history-server
COPY migrations /app/migrations
COPY config/unicorn-history-server/config.yml /app/config.yml

WORKDIR /app

ENTRYPOINT ["/app/unicorn-history-server"]
CMD ["--config", "config.yml"]
