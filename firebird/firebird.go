package firebird

/* for connecting to firebird and scan rdb$backup_history table*/

import (
	"database/sql"

	_ "github.com/nakagami/firebirdsql"
	"github.com/pharmacy72/gobak/backupitems"
)

const DRIVER = "firebirdsql"

type DatabaseApp struct {
	username string
	password string
	pathToDB string
}

//check existing last chain in backup_history
func (f *DatabaseApp) LastLastChainIntoFirebird(backupItems []*backupitems.BackupItem) (bool, error) {
	var n int
	conn, err := sql.Open(DRIVER, f.username+":"+f.password+"@127.0.0.1/"+f.pathToDB)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	for _, itm := range backupItems {

		stmt, err := conn.Prepare("SELECT Count(*) FROM rdb$backup_history where rdb$file_name = ?")
		defer stmt.Close()
		if err != nil {
			return false, err
		}

		row, err := stmt.Query(itm.FilePath())
		if err != nil {
			return false, err
		}
		for row.Next() {
			if err := row.Scan(&n); err != nil {
				return false, err
			}

			if n == 1 {
				return true, nil
			}
		}

	}

	return false, nil
}
