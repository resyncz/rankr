# build app from golang image
# we need golang image only for building an app
FROM golang:1.17.8-alpine3.15 as base_build

WORKDIR /app

COPY go.* .
RUN go mod download

COPY . .
RUN go build -v -o rankr-svc

# create runtime image from alpine
# we dont need fluffy entire golang image in final build
FROM alpine:3.15.0

RUN apk add ca-certificates

WORKDIR /usr/local/bin/rankr

COPY --from=base_build /app/rankr-svc .
COPY --from=base_build /app/conf/ ./conf/

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/rankr/rankr-svc"]