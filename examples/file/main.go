package main

import (
	"log"
	"os"

	//"github.com/underscorenygren/partaj/pkg/buffer"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/file"
	"github.com/underscorenygren/partaj/pkg/pipe"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("need to provide file argument")
	}

	src, err := file.NewSource(args[1])
	if err != nil {
		log.Fatal(err)
	}

	out := "./out.json"

	//sink := buffer.NewSink()
	sink, err := file.NewSink(out)
	if err != nil {
		log.Fatal(err)
	}

	p, err := pipe.NewStage(src, sink)
	err = p.Flow()

	if err != errors.ErrStreamEnd {
		log.Fatal(err)
	}

	sink.Close()

	//log.Printf("logged %d events", len(sink.Events))
}
