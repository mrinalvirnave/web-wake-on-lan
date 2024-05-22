# syntax=docker/dockerfile:1

FROM golang:1.22  AS build-stage

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY app/go.mod app/go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/reference/dockerfile/#copy
COPY app/ ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-wol

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /docker-wol /docker-wol

EXPOSE 5868

USER nonroot:nonroot

ENTRYPOINT ["/docker-wol"]


