# build stage
FROM golang:alpine AS build-env
RUN apk update && apk add git
ADD . /src
WORKDIR /src
ENV CGO_ENABLED 0
RUN go build \
    -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" \
    -o /tmp/bin
# run stage
FROM alpine
LABEL maintainer="dev@jpillora.com"
RUN apk update && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build-env /tmp/bin /app/bin
ENTRYPOINT ["/app/bin"]