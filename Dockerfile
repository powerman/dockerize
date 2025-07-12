FROM alpine:3.22.0

ARG TARGETARCH
ARG TARGETVARIANT
ARG GITHUB_REPOSITORY
ARG BINARY_NAME
ARG VERSION

ENV BINARY_PATH="/usr/local/bin/${BINARY_NAME}"

LABEL org.opencontainers.image.source="https://github.com/${GITHUB_REPOSITORY}"

# Copy the appropriate binary
COPY "dist/${BINARY_NAME}-${VERSION}-linux-${TARGETARCH}${TARGETVARIANT}" "${BINARY_PATH}"

# Make sure the binary is executable
RUN chmod +x "${BINARY_PATH}"

ENTRYPOINT ["${BINARY_PATH}"]
CMD ["--help"]
