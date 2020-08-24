package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type DBStorage struct {
	log *log.Logger
	db  *sql.DB
}

const (
	BALANCE_TABLE = "clientBalanceData"
	LOG_TABLE     = "clientLogData"
)

func SetupDB() (*DBStorage, error) {

	dbLogger := log.New(os.Stdout, "DBLog ", log.LstdFlags)
	dbLogger.Println("[TRACE] Setting up database")

	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", dbinfo)

	if err != nil {
		dbLogger.Println("[ERROR] connecting db", err)
		return nil, err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id Serial, balance Numeric)", BALANCE_TABLE))
	if err != nil {
		dbLogger.Printf("[ERROR] creating table %s. Reason: %s", BALANCE_TABLE, err)
		return nil, err
	}

	_, err = db.Exec(
		fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id Serial, sum Numeric, description VARCHAR, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW())",
			LOG_TABLE))
	if err != nil {
		dbLogger.Printf("[ERROR] creating table %s. Reason: %s", LOG_TABLE, err)
		return nil, err
	}

	dbLogger.Println("[TRACE] Database successfully created")

	return &DBStorage{
		log: dbLogger,
		db:  db,
	}, nil
}

func (ds *DBStorage) InsertData(data TableData) error {
	var err error
	if data.TableName == BALANCE_TABLE {
		_, err = ds.db.Exec(fmt.Sprintf("INSERT INTO %s VALUES (%d, 0)", data.TableName, data.ClientID))
	} else if data.TableName == LOG_TABLE {
		_, err = ds.db.Exec(fmt.Sprintf("INSERT INTO %s VALUES (%d, %d, '%s')",
			data.TableName, data.ClientID, data.Sum, data.Description))
	}
	return err
}

func (ds *DBStorage) UpdateData(data TableData) error {
	ds.log.Printf("[TRACE] Updating balance to %d for client with id %d", data.Sum, data.ClientID)
	_, err := ds.db.Exec(fmt.Sprintf("UPDATE %s SET balance = %d WHERE id = %d", data.TableName, data.Sum, data.ClientID))
	return err
}

func (ds *DBStorage) SelectData(data TableData) ([]TableData, error) {

	var query string
	if data.TableName == BALANCE_TABLE {
		query = fmt.Sprintf("SELECT balance FROM %s WHERE id = %d", data.TableName, data.ClientID)
	} else if data.TableName == LOG_TABLE {
		if data.Sort == "" || data.Sort == "time" {
			data.Sort = "created_at"
		}
		query = fmt.Sprintf("SELECT id, sum, description, created_at FROM %s WHERE id = %d ORDER BY %s DESC",
			data.TableName, data.ClientID, data.Sort)
	}
	rows, err := ds.db.Query(query)
	if err != nil {
		return []TableData{}, err
	}

	result := make([]TableData, 0)
	for rows.Next() {
		var selectedData TableData
		var err error

		if data.TableName == BALANCE_TABLE {
			err = rows.Scan(&selectedData.Sum)
		} else if data.TableName == LOG_TABLE {
			err = rows.Scan(&selectedData.ClientID, &selectedData.Sum, &selectedData.Description, &selectedData.Time)
		}

		if err != nil {
			return []TableData{}, err
		}

		result = append(result, selectedData)
	}

	if err := rows.Err(); err != nil {
		return []TableData{}, err
	}

	return result, err
}
