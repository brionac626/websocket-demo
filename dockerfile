FROM golang:1.12.5 As bs
LABEL maintainer="theone1632@gmail.com"

ENV GO111MODULE=on

COPY . ./websocketServer
RUN cd ./websocketServer && go build -o main

FROM alpine

WORKDIR /app
COPY --from=bs /go/websocketServer /app/
EXPOSE 8080
ENTRYPOINT [ "./main" ]