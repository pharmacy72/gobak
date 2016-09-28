package backupitems

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3" //justifying
)

var (
	once     sync.Once
	instance *sqliteRepository
)

type sqliteRepository struct {
	mu sync.Mutex
	db *sql.DB
}

func GetRepository() Repository {
	once.Do(func() {
		var err error
		instance = &sqliteRepository{}
		filedb := filepath.Join(filepath.Dir(os.Args[0]), "gobak.db")
		if _, e := os.Stat(filedb); e != nil && os.IsNotExist(e) {
			instance.db, err = sql.Open("sqlite3", filedb)
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
			_, err = instance.db.Exec(sqlStmt)
			if err != nil {
				panic(err)
			}
			log.Println("New reporsitory created.")
		} else {
			instance.db, err = sql.Open("sqlite3", filedb)
		}
		if err != nil {
			panic(err)
		}

	})
	return instance
}
func (s *sqliteRepository) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
func (s *sqliteRepository) Append(item *BackupItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	stmt, err := s.db.Prepare("INSERT INTO backup(GUID,GUID_PREV,LEVEL,DATE,STATUS,MD5,PathToFile,GUID_NEXT) values(?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(item.GUID, item.GUIDParent, item.Level, item.Date, item.Status, item.Hash, item.FileName, "_")
	if err != nil {
		return err
	}
	item.Insert = false
	item.Modified = false
	return nil
}

func (s *sqliteRepository) Update(item *BackupItem) error {
	stmt, err := s.db.Prepare("update backup set STATUS=?,PathToFile=?, MD5=? where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(item.Status, item.FileName, item.Hash, item.ID)
	if err != nil {
		return err
	}
	item.Modified = false
	return nil
}

func (s *sqliteRepository) Delete(item *BackupItem) error {
	//TODO:
	return nil
}

func (s *sqliteRepository) Refresh(item *BackupItem) error {
	//TODO:
	return nil
}

func (s *sqliteRepository) All() Collection {
	s.mu.Lock()
	defer s.mu.Unlock()
	return &sqliteCollection{
		rep: s,
		mu:  &s.mu,
	}
}
