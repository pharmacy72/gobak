package backupitems

import (
	"database/sql"
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
		instance.db, err = sql.Open("sqlite3", filedb)
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
func (s *sqliteRepository) Append(item BackupItem) error {
	return nil
}

func (s *sqliteRepository) Update(item BackupItem) error {
	return nil
}

func (s *sqliteRepository) Delete(item BackupItem) error {
	return nil
}

func (s *sqliteRepository) Refresh(item *BackupItem) error {
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
