FROM alpine:3.19

COPY build/event-collector /usr/local/bin/

ENTRYPOINT ["event-collector"]
CMD ["yhs.yml"]
