# Multi-stage build for the udp app
# Build stage
FROM golang:1.25-alpine AS build

WORKDIR /src/udp

# Install git for module downloads
RUN apk add --no-cache git

# Copy only udp module files
COPY ./go.mod ./go.sum ./
COPY . ./

# Download dependencies and build static binary
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /app/main .

# Final stage: minimal image
FROM scratch

# Copy the statically linked binary
COPY --from=build /app/main /app/main

EXPOSE 8080
ENTRYPOINT ["/app/main"]
