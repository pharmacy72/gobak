package dbase

import (
	"database/sql"
	"gobak/backupitems"
	"gobak/config"
	"gobak/level"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" //justifying
)

var db *sql.DB

func logerr(err error) {
	if err != nil {
		log.Println(err)
	}
}
func init() {
	var err error
	filedb := filepath.Join(filepath.Dir(os.Args[0]), "gobak.db")
	if _, e := os.Stat(filedb); e != nil && os.IsNotExist(e) {
		db, err = sql.Open("sqlite3", filedb)
		if err != nil {
			panic(err)
		}

		sqlStmt := `
			CREATE TABLE backup (
			 ID         INTEGER      PRIMARY KEY ASC AUTOINCREMENT NOT NULL,
   			 GUID       VARCHAR (38) NOT NULL,
   			 GUID_PREV  VARCHAR (38) NOT NULL,
   			 LEVEL      INTEGER      NOT NULL,
  			 DATE		DATE		 NOT NULL,


   			 STATUS     INTEGER,
   			 MD5        STRING       NOT NULL,
   			 PathToFile STRING       NOT NULL,
   			 GUID_NEXT  VARCHAR (38) NOT NULL
			);
			`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			panic(err)
		}
		log.Println("New reporsitory created.")
	} else {
		db, err = sql.Open("sqlite3", filedb)
	}
	if err != nil {
		panic(err)
	}
}

//FetchBackupItems it reads the data from the database about backups, returns them as an array
func FetchBackupItems() (res []*backupitems.BackupItem, err error) {
	x := make(map[string]*backupitems.BackupItem)
	rows, err := db.Query("SELECT ID,GUID,GUID_PREV,LEVEL, DATE ,STATUS,MD5,PathToFile,GUID_NEXT FROM backup order by date,level")
	if err != nil {
		return nil, err
	}

	defer func() {
		logerr(rows.Close())
	}()

	for rows.Next() {
		var (
			ID, inlevel, status                       int
			GUID, GUIDPrev, GUIDNext, MD5, pathToFile string
			date                                      time.Time
		)
		if err := rows.Scan(&ID, &GUID, &GUIDPrev, &inlevel, &date, &status, &MD5, &pathToFile, &GUIDNext); err != nil {
			return nil, err
		}
		item := backupitems.New(config.Current().PathToBackupFolder)
		item.ID = ID
		item.GUID = GUID
		item.GUIDParent = GUIDPrev
		item.Level = level.NewLevel(inlevel)
		item.Hash = MD5
		item.FileName = filepath.Base(pathToFile)
		item.Status = backupitems.StatusBackup(status)
		item.Insert = false
		item.Modified = false

		item.Date = date.Local()
		res = append(res, item)
		x[GUID] = item
	}

	for i := 0; i < len(res); i++ {
		res[i].Parent = x[res[i].GUIDParent]

	}
	return res, nil
}

//WriteBackupItem save data of BackupItem into a repository
func WriteBackupItem(item *backupitems.BackupItem) error {
	if item.Insert {
		stmt, err := db.Prepare("INSERT INTO backup(GUID,GUID_PREV,LEVEL,DATE,STATUS,MD5,PathToFile,GUID_NEXT) values(?,?,?,?,?,?,?,?)")
		if err != nil {
			return err
		}
		defer func() {
			logerr(stmt.Close())
		}()
		_, err = stmt.Exec(item.GUID, item.GUIDParent, item.Level, item.Date, item.Status, item.Hash, item.FileName, "_")
		if err != nil {
			return err
		}
	} else {
		if item.Modified {
			stmt, err := db.Prepare("update backup set STATUS=?,PathToFile=?, MD5=? where id=?")
			if err != nil {
				return err
			}
			defer func() {
				logerr(stmt.Close())
			}()
			_, err = stmt.Exec(item.Status, item.FileName, item.Hash, item.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//WriteBackupItems save data array of BackupItem into a repository
func WriteBackupItems(wrec []*backupitems.BackupItem) (err error) {
	for _, item := range wrec {
		if e := WriteBackupItem(item); e != nil {
			return e
		}
	}
	return nil
}
