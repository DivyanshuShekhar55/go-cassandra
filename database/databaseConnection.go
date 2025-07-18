package db

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

type DBconnection struct {
	Session *gocql.Session
}

var Connection DBconnection

func SetupDBConnection() {
	cluster := gocql.NewCluster("cassandra:9042")
	cluster.Keyspace = "chat"
	cluster.Consistency = gocql.Quorum
	Cs, err := cluster.CreateSession()
	Connection.Session = Cs

	if err != nil {
		fmt.Println("db err")
		panic(err)
	}

}

func ExecuteQuery(query string, args ...interface{}) error {
	err := Connection.Session.Query(query, args...).Exec()

	return err
}

func SelectQuery(query string, args ...interface{}) *gocql.Query {
	data := Connection.Session.Query(query, args...)
	return data
}

func getBucketForTime(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}

func ExecuteIterableQuery(query string, args ...interface{}) ([]any, error){
	var result []any
	var item any 

	iter := Connection.Session.Query(query, args).Iter()

	for iter.Scan(&item){
		tmp := item
		result = append(result, tmp)
	}

	if err := iter.Close(); err != nil {
        return nil, err
    }
    return result, nil
}
