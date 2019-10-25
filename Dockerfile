# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest

# Add Maintainer Info
LABEL maintainer="Yuriy Olkhovyy <y.olkhovyy@gmail.com>"

ENV GOPATH /go

# Set the Current Working Directory inside the container
WORKDIR /go/src/github.com/influxdata/telegraf


# Copy go mod and sum files
# COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
# RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . /go/src/github.com/influxdata/telegraf

# Build the Go app
RUN go get -u github.com/golang/dep/cmd/dep
RUN go get ./...
RUN make

# Expose port 8080 to the outside world
# EXPOSE 8080

# Command to run the executable
CMD ["./telegraf"]
