ARG BASE_IMAGE=alpine
ARG BASE_IMAGE_TAG=3.20
FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

COPY yunikorn-history-server /app/yunikorn-history-server
COPY migrations /app/migrations
COPY config.yml /app/config.yml

WORKDIR /app

ENTRYPOINT ["/app/yunikorn-history-server"]
CMD ["--config", "config.yml"]
