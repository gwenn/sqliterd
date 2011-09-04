/*
The author disclaims copyright to this source code.
*/
package main

import (
	"log"
	"os"
	"github.com/gwenn/sqlite"
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
		log.Fatalf("No database specified\n")
	}
	db, err := sqlite.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Error opening database '%s': %s\n", os.Args[1], err)
	}
	defer db.Close()
	tables, err := db.Tables()
	if err != nil {
		log.Fatalf("Error listing tables: %s\n", err)
	}
	entities := make([]*Entity, 0, 20)
	for _, table := range tables {
		columns, err := db.Columns(table)
		if err != nil {
			log.Fatalf("Error listing columns in '%s': %s\n", table, err)
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
			log.Fatalf("Error listing FKs in '%s': %s\n", table, err)
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
	template, err := template.ParseFile("tmpl.dot")
	if err != nil {
		log.Fatalf("Error loading template file: %s\n", err)
	}
	err = template.Execute(os.Stdout, erd)
	if err != nil {
		log.Fatalf("Error generating digraph: %s\n", err)
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
