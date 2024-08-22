FROM node:20 as builder

COPY web /build/src

WORKDIR /build/src

RUN npm install -g @angular/cli@17 && \
    npm install                    && \
    npm run build

FROM alpine:3.20

COPY --from=builder /build/assets /app/assets
COPY bin/app/yunikorn-history-server /app/yunikorn-history-server
COPY migrations /app/migrations
COPY config/yunikorn-history-server/config.yml /app/config.yml

WORKDIR /app

ENTRYPOINT ["/app/yunikorn-history-server"]
CMD ["--config", "config.yml"]
