# build stage
FROM golang:1.19 as build
LABEL maintainer="dev@jpillora.com"
ENV CGO_ENABLED 0
ADD . /src
WORKDIR /src
RUN go mod download
RUN go build \
    -ldflags "-X main.version=$(git describe --abbrev=0 --tags)" \
    -o serve
# run stage
FROM gcr.io/distroless/static-debian11
WORKDIR /app
CMD ["/app"]
COPY --from=build /src/serve /app/serve
ENTRYPOINT ["/app/serve"]