FROM golang:1.17.0

ADD . /usr/src/app
WORKDIR /usr/src/app

RUN go build -o hello-app
ENTRYPOINT [ "/usr/src/app/hello-app" ]
