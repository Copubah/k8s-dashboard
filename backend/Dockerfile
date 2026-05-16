FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/k8s-dashboard ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /bin/k8s-dashboard /k8s-dashboard
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/k8s-dashboard"]
