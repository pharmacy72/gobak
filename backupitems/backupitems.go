package backupitems

import (

	//"github.com/jinzhu/now"
	"gobak/level"
	"path/filepath"
	"strconv"
	"time"
)

//A StatusBackup type options for the items of backup
type StatusBackup int // bit mask

//A flags of StatusBackup
const (
	StatusNormal   StatusBackup = 0
	StatusArchived StatusBackup = 1
)

//A BackupItem contains information about a particular backup
type BackupItem struct {
	ID         int
	Level      level.Level
	GUID       string
	GUIDParent string
	Date       time.Time
	Status     StatusBackup
	Hash       string
	FileName   string
	Modified   bool
	Insert     bool
	Parent     *BackupItem
	basefolder string
}

//ChainWithAllParents returns an array with all the ancestors (levels)
// of the backup, including himself
func (r *BackupItem) ChainWithAllParents() (res []*BackupItem) {
	if r == nil {
		return nil
	}
	for n := r; n != nil; n = n.Parent {
		res = append([]*BackupItem{n}, res[:]...)
	}
	return res
}

//FilePath fullname on disk
func (r *BackupItem) FilePath() string {
	return filepath.Join(r.basefolder, strconv.Itoa(r.Level.Int()), r.FileName)
}

//IsArchived means that the backup is packed
func (r *BackupItem) IsArchived() bool {
	return r.Status&StatusArchived == StatusArchived
}

//New item of BackupItem
func New(pathbackup string) *BackupItem {
	result := &BackupItem{}
	result.basefolder = pathbackup
	return result
}
