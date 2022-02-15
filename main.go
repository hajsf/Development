package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	dbname = "postgres"
)

func main() {
	cmd := exec.Command("cmd", "pg_ctl", "status") // cmd := exec.Command("powershell", "pg_ctl", "status") status of the pgsql server
	/*	out, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		fmt.Printf("combined out:\n%s\n", string(out))
	*/
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("err at pgsql server status check")
		log.Printf("cmd.Run() failed with %s\n", err)
	}
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("\nout:\n%s\nerr:\n%s\n", outStr, errStr)

	cmd = exec.Command("powershell", "pg_ctl", "stop") // stop the pgsql server
	err = cmd.Run()
	if err != nil {
		fmt.Println("err at pgsql server shutdown", err)
	}

	cmd = exec.Command("powershell", "pg_ctl", "start") // start up the pgsql server

	err = cmd.Run()

	if err != nil {
		fmt.Println("err at pgsql server startup")
		log.Fatal(err)
	}

	// Terminal (not PC) Restart required for the system to capture the Env variables changes
	// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPSWD")

	println("user: ", user, ", password: ", password)

	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	println(psqlconn)

	// open database
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)

	// close database
	defer db.Close()

	// check db
	err = db.Ping()
	CheckError(err)

	fmt.Println("Connected!")

	sql := "CREATE TABLE IF NOT EXISTS people(id SERIAL PRIMARY KEY, name VARCHAR, salary INTEGER);"
	_, err = db.Exec(sql)
	if err != nil {
		fmt.Println("err1: ", err)
		panic(err)
	}

	// sql = "INSERT INTO people(name, salary) VALUES (@name, @salary);" // VALUES ($1, $2);
	//_, err = db.Exec(sql, "Joe", 10000)
	// https://www.calhoun.io/inserting-records-into-a-postgresql-database-with-gos-database-sql-package/
	// https://kb.objectrocket.com/postgresql/how-to-insert-record-in-postgresql-database-using-go-database-sql-package-785

	sqlStatement, err := db.Prepare("INSERT INTO people(name, salary) VALUES ($1, $2) RETURNING id;")
	if err != nil {
		fmt.Println("errx: ", err)
		log.Fatal(err)
	}
	defer sqlStatement.Close()
	idx := 0
	err = sqlStatement.QueryRow("Karam", 30000).Scan(&idx)
	if err != nil {
		fmt.Println("erry: ", err)
		log.Fatal(err)
	}
	fmt.Println("New record ID is:", idx)

	/*	sqlStatement := "INSERT INTO people(name, salary) VALUES ($1, $2) RETURNING id;"

		err = db.QueryRow(sqlStatement, "jon@calhoun.io", 300).Scan(&idx)
		if err != nil {
			fmt.Println("err2: ", err)
			panic(err)
		}
		fmt.Println("New record ID is:", idx)
	*/
	sqlQuery := `SELECT * FROM people`
	rows, err := db.Query(sqlQuery)
	if err != nil {
		fmt.Println("err3: ", err)
		panic(err)
	}
	// https://docs.immudb.io/1.0.0/jumpstart.html#sql-operations-with-the-go-sdk
	var (
		id     int
		name   string
		salary int
	)

	for rows.Next() {
		err := rows.Scan(&id, &name, &salary)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(id, name, salary)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
