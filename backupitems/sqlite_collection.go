package backupitems

import (
	"path/filepath"
	"time"

	"sync"

	"database/sql"

	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/level"
)

type sqliteCollection struct {
	mu  *sync.Mutex
	rep *sqliteRepository
}

func (c *sqliteCollection) doQuery() (*sql.Rows, error) {

	//TODO : use filters
	return c.rep.db.Query("SELECT ID,GUID,GUID_PREV,LEVEL, DATE ,STATUS,MD5,PathToFile,GUID_NEXT FROM backup order by date,level")
}
func (c *sqliteCollection) Get() (res []*BackupItem, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	x := make(map[string]*BackupItem)
	rows, err := c.doQuery()
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var (
			ID, inlevel, status                       int
			GUID, GUIDPrev, GUIDNext, MD5, pathToFile string
			date                                      time.Time
		)
		if err := rows.Scan(&ID, &GUID, &GUIDPrev, &inlevel, &date, &status, &MD5, &pathToFile, &GUIDNext); err != nil {
			return nil, err
		}
		item := New(config.Current().PathToBackupFolder)
		item.ID = ID
		item.GUID = GUID
		item.GUIDParent = GUIDPrev
		item.Level = level.NewLevel(inlevel)
		item.Hash = MD5
		item.FileName = filepath.Base(pathToFile)
		item.Status = StatusBackup(status)
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

func (c *sqliteCollection) AddFilterID(ids ...string) Collection {
	return c
}

func (c *sqliteCollection) ClearFilters() {

}
