FROM golang:1.15-buster

RUN apt-get update; apt-get install -fy nfs-ganesha nfs-ganesha-mem vim
RUN mkdir /nfs
COPY ganesha.conf /etc/ganesha
COPY run-tests.sh /

# Prepare the server
RUN mkdir /app
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o runtests cmd/main/runtests.go

EXPOSE 2049/tcp
ENTRYPOINT /run-tests.sh
