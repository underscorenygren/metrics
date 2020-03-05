package sql_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gosql "database/sql"
	"encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"github.com/underscorenygren/partaj/internal/logging"
	"github.com/underscorenygren/partaj/pkg/errors"
	"github.com/underscorenygren/partaj/pkg/sql"
	"github.com/underscorenygren/partaj/pkg/types"
	"go.uber.org/zap"
)

//Used to display test code in godoc
func Example() {}

var logger *zap.Logger

type Row struct {
	ID int `json:"id"`
}

func init() {
	logger = logging.ConfigureDevelopment(GinkgoWriter)
}

func scanner(rows *gosql.Rows) (*types.Event, error) {
	row := Row{}
	err := rows.Scan(&row.ID)
	Expect(err).To(BeNil())
	bytes, err := json.Marshal(row)
	logger.Debug("sql scan", zap.ByteString("bytes", bytes), zap.Int("id", row.ID))
	Expect(err).To(BeNil())
	e := types.NewEventFromBytes(bytes)
	return &e, nil
}

var _ = Describe("SQL", func() {

	dsn := "file:test.db?cache=shared&mode=memory"
	var source *sql.Source
	rowOne := Row{1}
	rowTwo := Row{2}

	BeforeEach(func() {
		db, err := gosql.Open("sqlite3", dsn)
		Expect(err).To(BeNil())
		_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY)")
		Expect(err).To(BeNil())

		_, err = db.Exec("INSERT INTO test_table VALUES (?), (?)", rowOne.ID, rowTwo.ID)
		Expect(err).To(BeNil())

		source, err = sql.NewSource(sql.SourceConfig{
			DB:     db,
			ScanFn: scanner,
			Stmt:   "SELECT * FROM test_table",
		})
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		Expect(source.Close()).To(BeNil())
	})

	It("reads events from sql", func(done Done) {
		row := Row{}

		e1, err := source.DrawOne()
		Expect(err).To(BeNil())

		logger.Info("read", zap.ByteString("e1", e1.Bytes()))
		err = json.Unmarshal(e1.Bytes(), &row)
		Expect(err).To(BeNil())
		Expect(row.ID).To(Equal(rowOne.ID))

		e2, err := source.DrawOne()
		Expect(err).To(BeNil())

		logger.Info("read", zap.ByteString("e2", e2.Bytes()))
		err = json.Unmarshal(e2.Bytes(), &row)
		Expect(err).To(BeNil())
		Expect(row.ID).To(Equal(rowTwo.ID))

		_, err = source.DrawOne()
		Expect(err).To(Equal(errors.ErrSQLEnd))

		close(done)
	})
})
