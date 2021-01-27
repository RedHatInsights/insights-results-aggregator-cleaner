package main

import (
	"fmt"
	"math"
	"time"

	"database/sql"

	_ "github.com/lib/pq"           // PostgreSQL database driver
	_ "github.com/mattn/go-sqlite3" // SQLite database driver

	"github.com/rs/zerolog/log"
)

// DBDriver type for db driver enum
type DBDriver int

const (
	// DBDriverSQLite3 shows that db driver is sqlite
	DBDriverSQLite3 DBDriver = iota
	// DBDriverPostgres shows that db driver is postgres
	DBDriverPostgres
	// DBDriverGeneral general sql(used for mock now)
	DBDriverGeneral
)

func initDatabaseConnection(configuration StorageConfiguration) (*sql.DB, error) {
	driverName := configuration.Driver
	dataSource := ""
	log.Info().Str("driverName", configuration.Driver).Msg("DB connection configuration")

	switch driverName {
	case "sqlite3":
		//driverType := DBDriverSQLite3
		//driver = &sqlite3.SQLiteDriver{}
		dataSource = configuration.SQLiteDataSource
	case "postgres":
		//driverType := DBDriverPostgres
		//driver = &pq.Driver{}
		dataSource = fmt.Sprintf(
			"postgresql://%v:%v@%v:%v/%v?%v",
			configuration.PGUsername,
			configuration.PGPassword,
			configuration.PGHost,
			configuration.PGPort,
			configuration.PGDBName,
			configuration.PGParams,
		)
	default:
		err := fmt.Errorf("driver %v is not supported", driverName)
		log.Err(err).Msg("Can not connect to data storage")
		return nil, err
	}

	connection, err := sql.Open(driverName, dataSource)

	if err != nil {
		log.Err(err).Msg("Can not connect to data storage")
		return nil, err
	}

	return connection, nil
}

func displayAllOldRecords(connection *sql.DB, maxAge string) error {
	query := "SELECT cluster, reported_at, last_checked_at FROM report WHERE reported_at < NOW() - $1::INTERVAL ORDER BY reported_at"
	rows, err := connection.Query(query, maxAge)
	if err != nil {
		return err
	}

	now := time.Now()

	for rows.Next() {
		var (
			clusterName string
			reported    time.Time
			lastChecked time.Time
		)

		if err := rows.Scan(&clusterName, &reported, &lastChecked); err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Unable to close the DB rows handle")
			}
			return err
		}

		age := int(math.Ceil(now.Sub(reported).Hours() / 24)) // in days
		log.Info().Str("cluster", clusterName).Str("reported", reported.Format(time.RFC3339)).Str("lastChecked", lastChecked.Format(time.RFC3339)).Int("age", age).Msg("Old report")
	}
	return nil
}
