ARG NODE_VERSION=20
ARG ALPINE_VERSION=3.20

FROM node:${NODE_VERSION} AS builder

COPY web /build/src

WORKDIR /build/src

RUN npm install -g @angular/cli@18 && \
    npm install                    && \
    npm run build

FROM alpine:${ALPINE_VERSION}

COPY --from=builder /build/assets /app/assets
COPY bin/app/unicorn-history-server /app/unicorn-history-server
COPY migrations /app/migrations
COPY config/unicorn-history-server/config.yml /app/config.yml

WORKDIR /app

ENTRYPOINT ["/app/unicorn-history-server"]
CMD ["--config", "config.yml"]
