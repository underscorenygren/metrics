package logging

import (
	"github.com/fgrosse/zaptest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
)

var zapper *zap.Logger
var atom zap.AtomicLevel = zap.NewAtomicLevel()

//Logger returns the globally configured logger
func Logger() *zap.Logger {
	return zapper
}

//All modules using this method rely on logger
//being present, so we error fatal if zap fails
//to init.
func init() {
	var err error
	config := zap.NewProductionConfig()
	config.Level = atom
	atom.SetLevel(zap.InfoLevel)
	zapper, err = config.Build()
	if err != nil {
		log.Fatal(err)
	}
}

//ConfigureDevelopment called by tests to signal that
//development config should be used
func ConfigureDevelopment(w io.Writer) *zap.Logger {
	zapper = zaptest.LoggerWriter(w)
	return zapper
}

//SetLevel sets level on global logger
func SetLevel(l zapcore.Level) {
	atom.SetLevel(l)
}
