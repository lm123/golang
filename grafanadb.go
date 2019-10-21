package main
import (
    "fmt"
    _ "database/sql"

    m "github.com/grafana/grafana/pkg/models"
    "github.com/go-xorm/xorm"
    _ "github.com/mattn/go-sqlite3"
)

func main() {
    connstr := fmt.Sprintf("file:/root/lm/grafana.db?cache=private&mode=rwc")
    engine, err := xorm.NewEngine("sqlite3", connstr)
    if err != nil {
	fmt.Printf("%s\n", err)
	return
    }

    engine.SetMaxOpenConns(0)
 
    var ids []int64
    err = engine.Table("dashboard").Cols("id").Find(&ids)

    if err != nil {
	fmt.Printf("%s\n", err)
	return
    }

    fmt.Println(ids)

    var dashboards = make([]*m.Dashboard, 0)
 
    engine.In("id", ids).Find(&dashboards)

    for _, value := range dashboards {
	fmt.Println(*value)
	//fmt.Println(value.Data)
    }

    new_connstr := fmt.Sprintf("file:/root/lm/new_grafana.db?cache=private&mode=rwc")
    new_engine, err := xorm.NewEngine("sqlite3", new_connstr)
    if err != nil {
	fmt.Printf("%s\n", err)
	return
    }

    for _, value := range dashboards {
        affected, err := new_engine.Insert(value)
        if err != nil {
	    fmt.Printf("%s\n", err)
	    return
        }
        fmt.Printf("affected %d records\n", affected)
    }
}
