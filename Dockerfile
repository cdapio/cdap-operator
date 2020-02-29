# Build the manager binary
FROM golang:1.13 as builder

# Copy in the go src
WORKDIR /go/src/cdap.io/cdap-operator
COPY api/ api/
COPY vendor/ vendor/
COPY main.go main.go
COPY controllers/ controllers/
COPY ./ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager ./main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /go/src/cdap.io/cdap-operator/manager .
COPY templates/ templates/
COPY config/crd/ crds/
COPY config/ config/
ENTRYPOINT ["/manager"]
