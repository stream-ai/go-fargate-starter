# Use the official Go image as the base image
FROM golang:1.22-alpine as builder

WORKDIR /workspace

# Copy go.mod and go.sum (dependency management)
RUN pwd
RUN tree
RUN ls -la
RUN ls -la /

COPY go.mod .
COPY go.sum .

RUN ls -la
RUN ls -la /

# Download project dependencies
RUN go mod download

# Copy source code
COPY . .
RUN tree

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /workspace/huddle-backend ./service/

FROM alpine:3.12
ARG TRAFFIC_PORT
COPY --from=builder /workspace/huddle-backend /app/huddle-backend
ENTRYPOINT ["/app/huddle-backend"]
EXPOSE ${TRAFFIC_PORT}
WORKDIR /app

