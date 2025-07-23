package db

import (
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
)

type DBconnection struct {
	Session *gocql.Session
}

var Connection DBconnection

func must(err error) {
	if err != nil {
		fmt.Println("ERROR WITH DB : ", err)
		log.Fatal(err)
	}
}

func SetupDBConnection() {
	cassHost := os.Getenv("CASSANDRA_HOST")
	if cassHost == "" {
		cassHost = "127.0.0.1"
	}

	cluster := gocql.NewCluster(cassHost)
	cluster.Keyspace = "chat"
	cluster.Consistency = gocql.Quorum
	Cs, err := cluster.CreateSession()
	must(err)
	Connection.Session = Cs

}

func ExecuteQuery(query string, args ...interface{}) error {
	err := Connection.Session.Query(query, args...).Exec()

	return err
}

func SelectQuery(query string, args ...interface{}) *gocql.Query {
	data := Connection.Session.Query(query, args...)
	return data
}

func ExecuteIterableQuery(query string, args ...interface{}) ([]any, error) {
	var result []any
	var item any

	iter := Connection.Session.Query(query, args).Iter()

	for iter.Scan(&item) {
		tmp := item
		result = append(result, tmp)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return result, nil
}
