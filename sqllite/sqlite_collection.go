package sqllite

import (
	"bytes"
	"database/sql"
	"github.com/pharmacy72/gobak/backupitems"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/level"
)

type sqliteCollection struct {
	mu        *sync.Mutex
	rep       *sqliteRepository
	filterIDs []string
}

func addFilter(filters []string, upper bool) (string, bool) {
	var res string
	for i, f := range filters {
		if upper {
			f = strings.ToUpper(f)
		}
		if i < len(filters)-1 {
			res += "\"" + f + "\","
		} else {
			res += "\"" + f + "\""
		}
	}
	return res, res != ""
}

func (c *sqliteCollection) doQuery() (*sql.Rows, error) {
	where := false
	var sqlbuf bytes.Buffer
	and := func() {
		if !where {
			sqlbuf.WriteString("\nwhere")
			where = true
		} else {
			sqlbuf.WriteString("\nand")
		}
	}

	sqlbuf.WriteString("SELECT ID,GUID,GUID_PREV,LEVEL, DATE ,STATUS,MD5,PathToFile,GUID_NEXT FROM backup")

	if ids, ok := addFilter(c.filterIDs, true); ok {
		and()
		sqlbuf.WriteString("\nUPPER(ID) in (" + ids + ")")
	}
	sqlbuf.WriteString("\norder by date,level")
	return c.rep.db.Query(sqlbuf.String())
}

//TODO: Create repository

func (c *sqliteCollection) Get() (res []*backupitems.BackupItem, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	x := make(map[string]*backupitems.BackupItem)
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

func (c *sqliteCollection) AddFilterID(ids ...string) backupitems.Collection {
	if len(ids) > 0 {
		c.filterIDs = append(c.filterIDs, ids...)
	}
	return c
}

func (c *sqliteCollection) ClearFilters() {
	c.filterIDs = make([]string, 0)
}
