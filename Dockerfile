FROM --platform=$BUILDPLATFORM golang:1.19.2-alpine AS build-env
ADD . /app
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o unwindia_ms_dotlan ./src

# Runtime image
FROM redhat/ubi8-minimal:8.7

RUN rpm -ivh https://dl.fedoraproject.org/pub/epel/epel-release-latest-8.noarch.rpm
RUN microdnf update && microdnf -y install ca-certificates inotify-tools

COPY --from=build-env /app/unwindia_ms_dotlan /
EXPOSE 8080
CMD ["./unwindia_ms_dotlan"]
