package backupitems

import (
	"path/filepath"
	"strconv"
)

//A StatusBackup type options for the items of backup
type StatusBackup int // bit mask

//A flags of StatusBackup
const (
	StatusNormal   StatusBackup = 0
	StatusArchived StatusBackup = 1
)

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

func New(pathBackup string) *BackupItem {
	return &BackupItem{
		basefolder: pathBackup,
	}
}
