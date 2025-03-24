package main

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"hereWeGoAgain/entities"
	"log"
	"reflect"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "gotraining"
)

func ExecuteQueryWithReturn(rows *sql.Rows, typoToReturn interface{}) any {
	t := reflect.TypeOf(typoToReturn).Elem()

	resultSet := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)

	for rows.Next() {
		entityReturned := reflect.New(t).Interface()

		dest := getScanDest(entityReturned)

		if err := rows.Scan(dest...); err != nil {
			log.Println("Error al escanear fila:", err)
			continue
		}

		resultSet = reflect.Append(resultSet, reflect.ValueOf(entityReturned).Elem())
	}

	return resultSet.Interface()
}

func getScanDest(entity interface{}) []interface{} {
	v := reflect.ValueOf(entity).Elem()

	dest := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		dest[i] = v.Field(i).Addr().Interface()
	}

	return dest
}

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable ", host, port, user, password, dbname)

	connector, err := pq.NewConnector(psqlInfo)

	db := sql.OpenDB(connector)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	txn, err := db.Begin()
	if err != nil {
		panic(err)
	}

	var entity entities.TestEntity

	entity.ID = 1
	entity.Name = "juancito"

	query := fmt.Sprintf("insert into public.test(id, name) values (%d, '%s') ON CONFLICT DO NOTHING", entity.ID, entity.Name)

	_, err = txn.Exec(query)
	if err != nil {
		panic(err)
	}

	query = fmt.Sprintf("select t.id, t.name from public.test t where id = %d ", entity.ID)

	result, err := txn.Query(query)
	if err != nil {
		panic(err)
	}

	resultRow := ExecuteQueryWithReturn(result, &entities.TestEntity{})

	resultSet, ok := resultRow.([]entities.TestEntity)
	if !ok {
		log.Fatal("Error al convertir resultado a []entities.TestEntity")
	} else {
		log.Println("Resultado al test:", resultSet)
	}

	for _, entit := range resultSet {
		fmt.Printf("La query devuelve: %d - %s\n", entit.ID, entit.Name)
	}

	err = txn.Commit()
	if err != nil {
		panic(err)
	}

}
