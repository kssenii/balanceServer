package handlers

import (
	"database/sql"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type DBStorage struct {
	log *log.Logger
	db  *sql.DB
}

func SetupDB() *DBStorage {

	dbLogger := log.New(os.Stdout, "DBLog ", log.LstdFlags)
	dbLogger.Println("[TRACE] Setting up database")

	connStr := "user=kseniia dbname=balancedb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		dbLogger.Println("[ERROR] connecting db", err)
		return nil
	}

	db.Exec("DROP TABLE IF EXISTS clientBalanceData")
	_, err = db.Exec("CREATE TABLE clientBalanceData  (id Serial, balance Numeric)")
	if err != nil {
		dbLogger.Println("[ERROR] creating table", err)
		return nil
	}

	dbLogger.Println("[TRACE] Database successfully created")

	return &DBStorage{
		log: dbLogger,
		db:  db,
	}
}

func (ds *DBStorage) InsertData(id uint32) error {
	_, err := ds.db.Exec("INSERT INTO clientBalanceData VALUES ($1, 0)", id)
	return err
}

func (ds *DBStorage) UpdateData(id uint32, sum uint64) error {
	ds.log.Printf("[TRACE] Updating balance to %d for client with id %d", sum, id)
	_, err := ds.db.Exec("UPDATE clientBalanceData SET balance = $2 WHERE id = $1", id, sum)
	return err
}

func (ds *DBStorage) SelectData(id uint32) (string, error) {

	rows, err := ds.db.Query("SELECT balance FROM clientBalanceData WHERE id = $1", id)
	if err != nil {
		return "", err
	}

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	data := strings.Join(names, ", ")
	log.Printf("[DEBUG] Selected data: %s", data)

	return data, err
}
