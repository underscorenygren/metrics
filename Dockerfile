from golang:latest

ENV GO111MODULE "on"

RUN mkdir -p /go/src/github.com/underscorenygren/partaj/
WORKDIR /go/src/github.com/underscorenygren/partaj/

#copy go mod files first
COPY go.mod .
COPY go.sum .

COPY . .

RUN make docker-install
RUN make build

EXPOSE 80

CMD ["./build/http"]
