package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/samber/lo"
)

var opts struct {
	Filename       string `short:"f" long:"filename" description:"CSV filename" required:"true"`
	TableName      string `short:"t" long:"table" description:"Table name" required:"false"`
	TrimWhiteSpace string `long:"trim" description:"Trim leading and trailing space" choice:"true" choice:"false" default:"true"`
}

func main() {
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		log.Fatalln(err)
	}

	var tableName string

	if opts.TableName == "" {
		csvFileName := filepath.Base(opts.Filename)
		tableName = strings.Split(csvFileName, ".")[0]
	} else {
		tableName = opts.TableName
	}

	trimWhiteSpace := opts.TrimWhiteSpace == "true"
	if err := invoke(tableName, opts.Filename, trimWhiteSpace); err != nil {
		log.Fatalln(err)
	}
}

func invoke(tableName string, filename string, trimWhilteSpace bool) error {

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	rawHeaders := data[0]

	headers := lo.Map(rawHeaders, func(value string, _ int) string {
		return EscapeColumnName(strings.TrimSpace(value))
	})

	dbFileName := "./db.db"
	db, err := sqlx.Connect("sqlite3", dbFileName)
	if err != nil {
		return err
	}
	defer db.Close()

	drop_table_statements := "DROP TABLE IF EXISTS " + tableName

	if _, err := db.Exec(drop_table_statements); err != nil {
		return err
	}

	create_table_statments := "CREATE TABLE " + tableName + " ("

	for idx, column := range headers {
		column := strings.TrimSpace(column)
		if len(headers) != idx+1 {
			create_table_statments += "\n  " + column + " varchar(256),"
		} else {
			create_table_statments += "\n  " + column + " varchar(256)"
		}
	}

	create_table_statments += "\n)"

	fmt.Println(create_table_statments)

	if _, err := db.Exec(create_table_statments); err != nil {
		return err
	}

	allRows := data[1:]
	for _, chunk := range lo.Chunk(allRows, 2) {
		statement := "INSERT INTO " + tableName + " ("
		statement += strings.Join(headers, ", ")
		statement += ") VALUES "

		rows := lo.Map(chunk, func(row []string, index int) string {
			values := lo.Map(row, func(value string, _ int) string {
				if trimWhilteSpace {
					value = strings.TrimSpace(value)
				}
				return "'" + EscapeStringValue(value) + "'"
			})

			return "(" + strings.Join(values, ", ") + ")"
		})

		statement += strings.Join(rows, ", ")

		fmt.Println(statement)

		if _, err := db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

func EscapeStringValue(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func EscapeColumnName(value string) string {
	return "\"" + strings.ReplaceAll(value, "\"", "\"\"") + "\""
}
