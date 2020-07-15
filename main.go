package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

// vote struct contains a single row from the votes table in the database.
// Each vote includes a artist and video id
type Vote struct {
	Artist   string
	VideoID  string
	UserName string
	TagID    string
}

// app struct contains global state.
type app struct {
	// db is the global database connection pool.
	db *sql.DB
}

// indexHandler handles requests to the / route.
func (app *app) indexHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if err := showTotals(w, r, app); err != nil {
			log.Printf("showTotals: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	case "POST":
		if err := saveVote(w, r, app); err != nil {
			log.Printf("saveVote: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	default:
		http.Error(w, fmt.Sprintf("HTTP Method %s Not Allowed", r.Method), http.StatusMethodNotAllowed)
	}
}

func main() {
	app := &app{}
	var err error
	// If the optional DB_TCP_HOST environment variable is set, it contains
	// the IP address and port number of a TCP connection pool to be created,
	// such as "127.0.0.1:5432". If DB_TCP_HOST is not set, a Unix socket
	// connection pool will be created instead.
	if os.Getenv("DB_TCP_HOST") != "" {
		app.db, err = initTCPConnectionPool()
		if err != nil {
			log.Fatalf("initTCPConnectionPool: unable to connect: %s", err)
		}
	} else {
		app.db, err = initSocketConnectionPool()
		if err != nil {
			log.Fatalf("initSocketConnectionPool: unable to connect: %s", err)
		}
	}

	// Create the votes table if it does not already exist.
	if _, err = app.db.Exec(`CREATE TABLE IF NOT EXISTS votes
	( vote_id SERIAL NOT NULL, time_cast timestamp NOT NULL,
	candidate CHAR(6) NOT NULL, PRIMARY KEY (vote_id) );`); err != nil {
		log.Fatalf("DB.Exec: unable to create table: %s", err)
	}

	http.HandleFunc("/", app.indexHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}

}

// showTotals renders json for total number of votes for artist.
func showTotals(w http.ResponseWriter, r *http.Request, app *app) error {
	// get total votes for each artist
	artist := r.URL.Query().Get("artist")
	if artist == "" {
		return fmt.Errorf("artist property missing from form submission")
	}
	videoID := r.URL.Query().Get("videoId")
	if videoID == "" {
		return fmt.Errorf("videoId property missing from form submission")
	}
	sqlSelect := "SELECT tag_id FROM votes WHERE artist_name=$1 AND video_id=$2 "
	rows, err := app.db.Query(sqlSelect, artist, videoID)

	voteCount := make(map[string]int)

	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			// do something with error
		} else {
			voteCount[tag]++
		}
	}
	js, err := json.Marshal(voteCount)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
	return nil
}

// saveVote saves a vote passed as http.Request form data.
func saveVote(w http.ResponseWriter, r *http.Request, app *app) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("Request.ParseForm: %v", err)
	}

	artist := r.FormValue("artist")
	if artist == "" {
		return fmt.Errorf("artist property missing from form submission")
	}

	videoID := r.FormValue("videoId")
	if videoID == "" {
		return fmt.Errorf("videoId property missing from form submission")
	}
	userName := r.FormValue("userName")
	if userName == "" {
		return fmt.Errorf("userName property missing from form submission")
	}
	tagID := r.FormValue("tagId")
	if tagID == "" {
		return fmt.Errorf("tagId property missing from form submission")
	}

	// [START cloud_sql_postgres_databasesql_connection]
	sqlInsert := "INSERT INTO votes(artist_name,video_id,user_name,tag_id)VALUES($1, $2, $3, $4)"

	if _, err := app.db.Exec(sqlInsert, artist, videoID, userName, tagID); err != nil {
		fmt.Fprintf(w, "unable to save vote: %s", err)
		return fmt.Errorf("DB.Exec: %v", err)
	} else {
		fmt.Fprintf(w, "Vote successfully cast for %s!\n", artist)
	}

	return nil
	// [END cloud_sql_postgres_databasesql_connection]
}

// mustGetEnv is a helper function for getting environment variables.
// Displays a warning if the environment variable is not set.
func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("Warning: %s environment variable not set.\n", k)
	}
	return v
}

// initSocketConnectionPool initializes a Unix socket connection pool for
// a Cloud SQL instance of SQL Server.
func initSocketConnectionPool() (*sql.DB, error) {
	// [START cloud_sql_postgres_databasesql_create_socket]
	var (
		dbUser                 = mustGetenv("DB_USER")
		dbPwd                  = mustGetenv("DB_PASS")
		instanceConnectionName = mustGetenv("INSTANCE_CONNECTION_NAME")
		dbName                 = mustGetenv("DB_NAME")
	)

	socketDir, isSet := os.LookupEnv("DB_SOCKET_DIR")
	if !isSet {
		socketDir = "/cloudsql"
	}

	var dbURI string
	dbURI = fmt.Sprintf("user=%s password=%s database=%s host=%s/%s", dbUser, dbPwd, dbName, socketDir, instanceConnectionName)

	// dbPool is the pool of database connections.
	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	// [START_EXCLUDE]
	configureConnectionPool(dbPool)
	// [END_EXCLUDE]

	return dbPool, nil
	// [END cloud_sql_postgres_databasesql_create_socket]
}

// initTCPConnectionPool initializes a TCP connection pool for a Cloud SQL
// instance of SQL Server.
func initTCPConnectionPool() (*sql.DB, error) {
	// [START cloud_sql_postgres_databasesql_create_tcp]
	var (
		dbUser    = mustGetenv("DB_USER")
		dbPwd     = mustGetenv("DB_PASS")
		dbTcpHost = mustGetenv("DB_TCP_HOST")
		dbPort    = mustGetenv("DB_PORT")
		dbName    = mustGetenv("DB_NAME")
	)

	var dbURI string
	dbURI = fmt.Sprintf("host=%s user=%s password=%s port=%s database=%s", dbTcpHost, dbUser, dbPwd, dbPort, dbName)

	// dbPool is the pool of database connections.
	dbPool, err := sql.Open("pgx", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}

	// [START_EXCLUDE]
	configureConnectionPool(dbPool)
	// [END_EXCLUDE]

	return dbPool, nil
	// [END cloud_sql_postgres_databasesql_create_tcp]
}

// configureConnectionPool sets database connection pool properties.
// For more information, see https://golang.org/pkg/database/sql
func configureConnectionPool(dbPool *sql.DB) {
	// [START cloud_sql_postgres_databasesql_limit]

	// Set maximum number of connections in idle connection pool.
	dbPool.SetMaxIdleConns(5)

	// Set maximum number of open connections to the database.
	dbPool.SetMaxOpenConns(7)

	// [END cloud_sql_postgres_databasesql_limit]

	// [START cloud_sql_postgres_databasesql_lifetime]

	// Set Maximum time (in seconds) that a connection can remain open.
	dbPool.SetConnMaxLifetime(1800)

	// [END cloud_sql_postgres_databasesql_lifetime]
}
