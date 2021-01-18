FROM golang:buster

WORKDIR /src
COPY . /src
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o cloudsort .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /src/cloudsort .
ENTRYPOINT ["/root/cloudsort"]  
