package database

import (
	"database/sql"
	"embed"
	"fmt"
	"sync"

	C "github.com/libost/bandori-tg/constants"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

var (
	db     *sql.DB
	dbOnce sync.Once
	dbErr  error
)

func getDB() (*sql.DB, error) {
	dbOnce.Do(func() {
		db, dbErr = sql.Open("sqlite", C.DatabaseFile+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)")
		if dbErr != nil {
			return
		}

		schema, err := schemaFS.ReadFile("schema.sql")
		if err != nil {
			dbErr = err
			return
		}

		if _, err := db.Exec(string(schema)); err != nil {
			dbErr = err
			return
		}
	})
	if dbErr != nil {
		return nil, dbErr
	}
	return db, nil
}

func initCase() (map[string]any, error) {
	data := map[string]any{}
	return data, nil
}

func createUserIfNotExists(conn *sql.DB, id int64) error {
	_, err := conn.Exec(
		"INSERT OR IGNORE INTO USERPOOL (user_id) VALUES (?)",
		id,
	)
	return err
}

func createCase(id int64, conn *sql.DB) (map[string]any, error) {
	data := map[string]any{"user_id": id, "exists": false}
	if err := createUserIfNotExists(conn, id); err != nil {
		return nil, err
	}
	data["exists"] = true
	return data, nil
}

func langCase(id int64, conn *sql.DB) (map[string]any, error) {
	var displayLang, queryLang string
	err := conn.QueryRow(
		"SELECT COALESCE(display_language, ''), COALESCE(query_language, '') FROM USERPOOL WHERE user_id = ?",
		id,
	).Scan(&displayLang, &queryLang)
	if err != nil {
		return nil, err
	}
	return map[string]any{"display_language": displayLang, "query_language": queryLang}, nil
}

func setDisplayLanguageCase(id int64, langCode string, conn *sql.DB) (map[string]any, error) {
	_, err := conn.Exec(
		"UPDATE USERPOOL SET display_language = ? WHERE user_id = ?",
		langCode, id,
	)
	if err != nil {
		return nil, err
	}
	return map[string]any{"display_language": langCode}, nil
}

func setQueryLanguageCase(id int64, langCode string, conn *sql.DB) (map[string]any, error) {
	_, err := conn.Exec(
		"UPDATE USERPOOL SET query_language = ? WHERE user_id = ?",
		langCode, id,
	)
	if err != nil {
		return nil, err
	}
	return map[string]any{"query_language": langCode}, nil
}

func setAdminCase(id int64, conn *sql.DB) (map[string]any, error) {
	_, err := conn.Exec(
		"UPDATE USERPOOL SET user_group = 'admin' WHERE user_id = ?",
		id,
	)
	if err != nil {
		return nil, err
	}
	return map[string]any{"user_group": "admin"}, nil
}

func getUserGroupCase(id int64, conn *sql.DB) (map[string]any, error) {
	var userGroup string
	err := conn.QueryRow(
		"SELECT COALESCE(user_group, 'user') FROM USERPOOL WHERE user_id = ?",
		id,
	).Scan(&userGroup)
	if err != nil {
		return nil, err
	}
	return map[string]any{"user_group": userGroup}, nil
}

func Init(request string, id int64, other map[string]any) (map[string]any, error) {
	conn, err := getDB()
	if err != nil {
		return nil, err
	}

	switch request {
	case "init":
		return initCase()
	case "create":
		return createCase(id, conn)
	case "lang":
		return langCase(id, conn)
	case "set_display_language":
		langCode, ok := other["lang_code"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid lang_code type")
		}
		return setDisplayLanguageCase(id, langCode, conn)
	case "set_query_language":
		langCode, ok := other["lang_code"].(string)
		if !ok {
			return nil, fmt.Errorf("Invalid lang_code type")
		}
		return setQueryLanguageCase(id, langCode, conn)
	case "setadmin":
		return setAdminCase(id, conn)
	case "get_user_group":
		return getUserGroupCase(id, conn)
	default:
		return nil, fmt.Errorf("Unknown request: %s", request)
	}
}
