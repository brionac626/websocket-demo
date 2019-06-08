FROM golang:1.12.5 As bs
LABEL maintainer="theone1632@gmail.com"

ENV GO111MODULE=on

COPY . ./websocketServer
RUN cd ./websocketServer && CGO_ENABLED=0 go build \ 
    -installsuffix 'static' \
    -o main

FROM alpine:latest

# RUN apk add --no-cache bash
WORKDIR /app
COPY --from=bs /go/websocketServer/main /app/
EXPOSE 8080
ENTRYPOINT [ "./main" ]