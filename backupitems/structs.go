package backupitems

import (
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/level"
	"github.com/pharmacy72/gobak/md5f"
	"github.com/pharmacy72/gobak/zip"
	"time"
)

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
	md5App     *md5f.Md5App
	zip        *zip.CompressApp
	fileutils  *fileutils.FileUtils
}
