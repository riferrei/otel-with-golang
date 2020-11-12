FROM golang:1.15.2

ADD . /usr/src/app
WORKDIR /usr/src/app

RUN go build -o hello-app
ENTRYPOINT [ "/usr/src/app/hello-app" ]
