package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kimxilxyong/gorp"
	_ "github.com/lib/pq"
	"log"
	"math/rand"
	"os"
	//"reflect"
	"unicode"
)

// Print Debug info to stdout (0: off, 1: error, 2: warning, 3: info, 4: debug)
var DebugLevel int = 3

type testable interface {
	GetId() int64
	Rand()
}

// See: https://github.com/go-gorp/gorp/issues/175
type AliasTransientField struct {
	Id     int64  `db:"id"`
	Bar    int64  `db:"-"`
	BarStr string `db:"bar"`
}

func (me *AliasTransientField) GetId() int64 { return me.Id }
func (me *AliasTransientField) Rand() {
	me.BarStr = fmt.Sprintf("random %d", rand.Int63())
}

func TestGorp() (err error) {
	//drivername := "postgres"
	//dsn := "user=gorptest password=gorptest dbname=gorptest sslmode=disable"
	//dialect := gorp.PostgresDialect{}

	drivername := "mysql"
	dsn := "gorptest:gorptest@/gorptest?parseTime=true"
	dialect := gorp.MySQLDialect{"InnoDB", "UTF8"}

	// connect to db using standard Go database/sql API
	db, err := sql.Open(drivername, dsn)
	if err != nil {
		return errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data:
	if err = db.Ping(); err != nil {
		return errors.New("db.Ping failed: " + err.Error())
	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
	defer dbmap.Db.Close()
	dbmap.DebugLevel = DebugLevel
	// Will log all SQL statements + args as they are run
	// The first arg is a string prefix to prepend to all log messages
	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "testgorp:", log.Lmicroseconds))

	foo := AliasTransientField{BarStr: "Foo: some BarStr with 'quotes' in it"}

	// register the structs you wish to use with gorp
	// you can also use the shorter dbmap.AddTable() if you
	// don't want to override the table name
	// SetKeys(true) means we have a auto increment primary key, which
	// will get automatically bound to your struct post-insert
	table := dbmap.AddTable(foo)
	table.SetKeys(true, "id")

	// create the table. in a production system you'd generally
	// use a migration tool, or create the tables via scripts
	if err = dbmap.CreateTablesIfNotExists(); err != nil {
		return errors.New("Create tables failed: " + err.Error())
	}

	// Force create all indexes for this database
	if err = dbmap.CreateIndexes(); err != nil {
		return errors.New("Create indexes failed: " + err.Error())
	}

	if err = dbmap.Insert(&foo); err != nil {
		return errors.New("Insert failed: " + err.Error())
	}

	insertsql := `insert into aliastransientfield ("id", "bar") values(default, ?)`
	_, err = dbmap.Exec(insertsql, "test insert \"double quote\", 'single quote', `backtick`")
	if err != nil {
		return errors.New("Insert failed: " + err.Error())
	}

	insertsql = `insert into aliastransientfield ("id", "bar") values(default, :barstring)`
	_, err = dbmap.Exec(insertsql, map[string]interface{}{
		"barstring": "test named insert \"double quote\", 'single quote', `backtick`"})
	if err != nil {
		return errors.New("Named Insert failed: " + err.Error())
	}

	fmt.Println("--------------- STARTING SELECT -----------------")

	// Select *
	//foobar := &AliasTransientField{Id: 1, BarStr: "some BarStr with 'quotes' in it"}

	//var foos []AliasTransientField
	rows, err := dbmap.Select(foo, "select * from "+table.TableName)

	if err != nil {
		return errors.New(fmt.Sprintf("couldn't select * from %s err=%v", table.TableName, err))
	} else if len(rows) < 1 {
		return errors.New(fmt.Sprintf("unexpected row count in %s: %d", table.TableName, len(rows)))
		//} else if !reflect.DeepEqual(foo, rows[0]) {
		//	return errors.New(fmt.Sprintf("select * result: %v != %v", foo, rows[0]))
	}
	for _, row := range rows {

		af := row.(*AliasTransientField)
		fmt.Printf("ID: %d, BarStr: %s\n", af.GetId(), af.BarStr)
	}

	//fmt.Printf("Id: %d\n", rows[0].(Id))

	/*
		var bar string
		for rows.Next() {
			err = rows.Scan(&bar)
		}
	*/
	return
}

func stringMinifier(in string) (out string) {

	white := false
	for _, c := range in {
		if unicode.IsSpace(c) {
			if !white {
				out = out + " "
			}
			white = true
		} else {
			out = out + string(c)
			white = false
		}
	}

	return
}

func main() {
	err := TestGorp()
	if err != nil {

		panic(err)

	}
}
