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
	Id     int64                      `db:"id"`
	Bar    int64                      `db:"-"`
	BarStr string                     `db:"bar"`
	Childs []AliasTransientFieldChild `db:"relation:ParentId"` // will create a table Comment as a detail table with foreignkey PostId
	// if you want a different name just issue a: table = dbmap.AddTableWithName(post.Comment{}, "comments_embedded_test")
	// after: table := dbmap.AddTableWithName(post.Post{}, "posts_embedded_test")
	// but before: dbmap.CreateTablesIfNotExists()
}

type AliasTransientFieldChild struct {
	Id       int64  `db:"id"`
	ParentId uint64 `db:"notnull, index:idx_foreign_key_parentid"` // points to post.id

	Bar    int64  `db:"-"`
	BarStr string `gorp:"childbar"`
}

func (me *AliasTransientField) GetId() int64 { return me.Id }
func (me *AliasTransientField) Rand() {
	me.BarStr = fmt.Sprintf("random %d", rand.Int63())
}

func TestGorp() (err error) {
	//drivername := "postgres"
	//dsn := "user=golang password=golang dbname=golang sslmode=disable"
	//dialect := gorp.PostgresDialect{}

	drivername := "mysql"
	dsn := "golang:golang@/golang?parseTime=true&clientFoundRows=true"
	dialect := gorp.MySQLDialect{"InnoDB", "utf8mb4"}

	// connect to db using standard Go database/sql API
	db, err := sql.Open(drivername, dsn)
	defer db.Close()
	if err != nil {
		return errors.New("sql.Open failed: " + err.Error())
	}

	// Open doesn't open a connection. Validate DSN data using ping
	if err = db.Ping(); err != nil {
		return errors.New("db.Ping failed: " + err.Error())
	}

	// Set the connection to use utf8bmb4
	if dialect.Engine == "InnoDB" {
		fmt.Println("Setting connection to utf8mb4")
		_, err = db.Exec("SET NAMES utf8mb4 COLLATE utf8mb4_general_ci")
		if err != nil {
			return errors.New("SET NAMES utf8mb4 COLLATE utf8mb4_general_ci: " + err.Error())
		}
	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: dialect}
	//defer dbmap.Db.Close()
	dbmap.DebugLevel = DebugLevel
	dbmap.CheckAffectedRows = true

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

	tablechild := dbmap.AddTable(AliasTransientFieldChild{})
	tablechild.SetKeys(true, "id")

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

	var insertsql string
	if drivername == "postgres" {
		insertsql = `insert into aliastransientfield ("id", "bar") values(default, $1)`

	} else {
		insertsql = "insert into AliasTransientField (`id`,`bar`) values (null,?)"
	}
	_, err = dbmap.Exec(insertsql, "test insert \"double quote\", 'single quote', `backtick`")
	if err != nil {
		return errors.New("Insert failed: " + err.Error())
	}

	insertsql = "insert into aliastransientfield (id, bar) values(default, :barstring)"
	_, err = dbmap.Exec(insertsql, map[string]interface{}{
		"barstring": "test named insert \"double quote\", 'single quote', `backtick`"})
	if err != nil {
		return errors.New("Named Insert failed: " + err.Error())
	}

	fmt.Println("--------------- STARTING SELECT -----------------")

	// Method 1: Results are returned as an array
	// of interfaces (=rows here)
	rows, err := dbmap.Select(foo, "select * from "+table.TableName)
	if err != nil {
		return errors.New(fmt.Sprintf("couldn't select * from %s err=%v", table.TableName, err))
	} else if len(rows) < 1 {
		return errors.New(fmt.Sprintf("unexpected row count in %s: %d", table.TableName, len(rows)))
	}
	// Method 1: read the rows from the returned array of interfaces ( rows := []interface{} )
	for _, row := range rows {

		// cast the row to our struct
		af := row.(*AliasTransientField)
		fmt.Printf("Method1: ID: %d, BarStr: %s\n", af.GetId(), af.BarStr)
	}

	// Method 2: Resulting rows are appended to a pointer of a slice (foos)
	var foos []AliasTransientField
	_, err = dbmap.Select(&foos, "select * from "+table.TableName)
	if err != nil {
		return errors.New(fmt.Sprintf("couldn't select * from %s err=%v", table.TableName, err))
	} else if len(foos) < 1 {
		return errors.New(fmt.Sprintf("unexpected row count in %s: %d", table.TableName, len(foos)))
	}
	// Method 2: read the rows from the input slice (=foos)
	for _, f := range foos {
		fmt.Printf("Method2: ID: %d, BarStr: %s\n", f.Id, f.BarStr)
	}

	// Test embedded child insert/update
	foo = AliasTransientField{BarStr: "Parent 2"}

	child := AliasTransientFieldChild{BarStr: "1Some xxx child Foo1: some BarStr with 'quotes' in it"}
	foo.Childs = append(foo.Childs, child)

	child = AliasTransientFieldChild{BarStr: "2Some xxxxx child Foo2: some BarStr with 'quotes' in it"}
	foo.Childs = append(foo.Childs, child)

	var rowcount int64 = 0
	rowcount, err = dbmap.UpdateWithChilds(&foo)
	//err = dbmap.InsertWithChilds(&foo)
	if err != nil {
		return errors.New("UpdateWithChilds failed: " + err.Error())
	}
	fmt.Printf("UpdateWithChilds %d\n", rowcount)
	for _, c := range foo.Childs {
		fmt.Printf("Child: ID: %d, Parent: %d, Str: %s\n", c.Id, c.ParentId, c.BarStr)
	}
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
