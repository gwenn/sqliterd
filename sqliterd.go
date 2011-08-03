/*
The author disclaims copyright to this source code.
*/
package main

import (
	"fmt"
	"os"
	sqlite "github.com/gwenn/sqlite"
	"template"
)

type Erd struct {
	Entities  []*Entity
	Relations []*Relationship
}

type Entity struct {
	Name       string
	Attributes []*Attribute
}

type Attribute struct {
	Name string
	Key  bool
}

type Relationship struct {
	Child     string
	ChildKey  string
	Parent    string
	ParentKey string
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "No database specified\n")
		os.Exit(1)
	}
	db, err := sqlite.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database '%s': %s\n", os.Args[1], err)
		os.Exit(1)
	}
	defer db.Close()
	tables, err := db.Tables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing tables: %s\n", err)
		os.Exit(1)
	}
	entities := make([]*Entity, 0, 20)
	for _, table := range tables {
		columns, err := db.Columns(table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing columns in '%s': %s\n", table, err)
			os.Exit(1)
		}
		attrs := make([]*Attribute, len(columns))
		for i, col := range columns {
			attrs[i] = &Attribute{Name: col.Name}
		}
		entity := Entity{table, attrs}
		entities = append(entities, &entity)
	}
	rs := make([]*Relationship, 0, 20)
	for _, table := range tables {
		fks, err := db.ForeignKeys(table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing FKs in '%s': %s\n", table, err)
			os.Exit(1)
		}
		if len(fks) > 0 {
			for _, fk := range fks {
				for _, col := range fk.From {
					updateKey(entities, table, col)
				}
				for _, col := range fk.To {
					updateKey(entities, fk.Table, col)
				}
				// TODO How to handle composite keys?
				rs = append(rs, &Relationship{table, fk.From[0], fk.Table, fk.To[0]})
			}
		}
	}
	render(&Erd{entities, rs})
}

func render(erd *Erd) {
	template, err := template.ParseFile("tmpl.dot", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading template file: %s\n", err)
		os.Exit(1)
	}
	err = template.Execute(os.Stdout, erd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating digraph: %s\n", err)
		os.Exit(1)
	}
}

func updateKey(entities []*Entity, table string, col string) {
	for _, entity := range entities {
		if entity.Name == table {
			for _, attr := range entity.Attributes {
				if attr.Name == col {
					attr.Key = true
				}
			}
		}
	}
}
