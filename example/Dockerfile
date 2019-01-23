# Base build image
FROM golang:1.11 AS build_base
 
# Install some dependencies needed to build the project
WORKDIR /go/src
 
# Force the go compiler to use modules
ENV GO111MODULE=on
 
# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
 
#This is the ‘magic’ step that will download all the dependencies that are specified in 
# the go.mod and go.sum file.
# Because of how the layer caching system works in Docker, the  go mod download 
# command will _ only_ be re-run when the go.mod or go.sum file change 
# (or when we add another docker instruction this line)
RUN go mod download
 
# This image builds the weavaite server
FROM build_base AS server_builder
# Here we copy the rest of the source code
COPY . .
# And compile the project
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o test-command -a -tags netgo -ldflags '-w -extldflags "-static"' ./cmd/test-command
 
#In this last stage, we start from a fresh Alpine image, to reduce the image size and not ship the Go compiler in our production artifacts.
FROM alpine AS testcommand
# We add the certificates to be able to verify remote weaviate instances
RUN apk add ca-certificates
# Finally we copy the statically compiled Go binary.
COPY --from=server_builder /go/src/test-command /bin/test-command
ENTRYPOINT ["/bin/test-command"]
 