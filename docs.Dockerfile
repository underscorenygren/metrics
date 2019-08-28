from golang:latest

RUN apt-get update --fix-missing && apt-get install -y golang-doc

RUN mkdir -p /go/src/github.com/underscorenygren/partaj/

RUN go get golang.org/x/tools/cmd/godoc

COPY . /go/src/github.com/underscorenygren/partaj/

EXPOSE 6060

CMD ["/go/bin/godoc",  "-http=:6060"]
