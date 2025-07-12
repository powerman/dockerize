FROM alpine:3.22.0

ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION

LABEL org.opencontainers.image.source="https://github.com/powerman/dockerize"

# Copy the appropriate binary
COPY ".cache/dist/dockerize-${VERSION}-linux-${TARGETARCH}${TARGETVARIANT}" /usr/local/bin/dockerize

# Make sure the binary is executable
RUN chmod +x /usr/local/bin/dockerize

ENTRYPOINT ["/usr/local/bin/dockerize"]
CMD ["--help"]
