FROM golang:1.24 as builder

WORKDIR /workspace

ARG GOPROXY
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN GOPROXY=$GOPROXY go mod download

# Copy the go source
COPY main.go main.go
COPY pkg/ pkg/

# Build -mod=vendor
RUN CGO_ENABLED=0 go build -a -o ks-extension-upgrade

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/ks-extension-upgrade .
USER 65532:65532

CMD ["/ks-extension-upgrade", "--kubeconfig", "kube.config"]