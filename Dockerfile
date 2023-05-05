# Build the manager binary
FROM golang:1.19 as builder

# Copy everything in the go src
WORKDIR /go/src/cdap.io/cdap-operator
COPY ./ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager ./main.go

# Copy the controller-manager into a thin image
FROM ubuntu:latest
RUN apt-get update --allow-releaseinfo-change && apt-get upgrade -y
WORKDIR /
COPY --from=builder /go/src/cdap.io/cdap-operator/manager .
COPY templates/ templates/
COPY config/ config/
ENTRYPOINT ["/manager"]
