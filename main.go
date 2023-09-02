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

type Options struct {
	Filename       string `short:"f" long:"filename" description:"CSV filename" required:"true"`
	TableName      string `short:"t" long:"table" description:"Table name" required:"false"`
	TrimWhiteSpace string `long:"trim" description:"Trim leading and trailing space" choice:"true" choice:"false" default:"true"`
	DataTypes      string `long:"data_types" description:"Data types of the fields" required:"false"`
	ChunkSize      int    `long:"chunk_size" description:"Chunk size" default:"10"`
}

type Config struct {
	DefaultDataType string
	TrimWhiteSpace  bool
	ChunkSize       int
}

func main() {
	var opts = Options{}
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

	config := Config{
		DefaultDataType: opts.DataTypes,
		TrimWhiteSpace:  strings.ToLower(opts.TrimWhiteSpace) == "true",
		ChunkSize:       opts.ChunkSize,
	}

	if err := invoke(tableName, opts.Filename, config); err != nil {
		log.Fatalln(err)
	}
}

func invoke(tableName string, filename string, config Config) error {

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
		var dataType string

		if len(strings.TrimSpace(config.DefaultDataType)) > 0 {
			dataType = " " + config.DefaultDataType
		} else {
			dataType = ""
		}

		if len(headers) != idx+1 {
			create_table_statments += fmt.Sprintf("\n %s%s,", column, dataType)
		} else {
			create_table_statments += fmt.Sprintf("\n %s%s", column, dataType)
		}
	}

	create_table_statments += "\n)"

	fmt.Println(create_table_statments)

	if _, err := db.Exec(create_table_statments); err != nil {
		return err
	}

	allRows := data[1:]
	for _, chunk := range lo.Chunk(allRows, config.ChunkSize) {
		statement := "INSERT INTO " + tableName + " ("
		statement += strings.Join(headers, ", ")
		statement += ") VALUES "

		rows := lo.Map(chunk, func(row []string, index int) string {
			values := lo.Map(row, func(value string, _ int) string {
				if config.TrimWhiteSpace {
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
