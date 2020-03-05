package sql

import (
	"database/sql"
	"fmt"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//ScanFn  method signature that defines how a row is processed
type ScanFn func(*sql.Rows) (*types.Event, error)

//SourceConfig input to NewSource()
type SourceConfig struct {
	DB     *sql.DB
	ScanFn ScanFn
	Stmt   string
}

//Source implements source interface for sql
type Source struct {
	rows *sql.Rows
	cfg  SourceConfig
}

//NewSource craetes a new source that reads events from sql
func NewSource(conf SourceConfig) (*Source, error) {
	if conf.DB == nil {
		return nil, fmt.Errorf("no db provided")
	}

	if conf.ScanFn == nil {
		return nil, fmt.Errorf("no scan fn provided")
	}

	return &Source{
		cfg:  conf,
		rows: nil,
	}, nil
}

//DrawOne reads one event from DB, potentially running query
func (source *Source) DrawOne() (*types.Event, error) {
	logger := logging.Logger()

	if source.rows == nil {
		stmt := source.cfg.Stmt
		logger.Debug("querying", zap.String("q", stmt))
		rows, err := source.cfg.DB.Query(stmt)
		if err != nil {
			return nil, err
		}
		source.rows = rows
	}

	if !source.rows.Next() {
		return nil, errors.ErrSQLEnd
	}

	return source.cfg.ScanFn(source.rows)
}

//Close closes db connection ,and any rows used
func (source *Source) Close() error {
	if source.rows != nil {
		return source.rows.Close()
	}
	source.cfg.DB.Close()
	return nil
}
