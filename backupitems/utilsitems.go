package backupitems

import (
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/md5f"
	"github.com/pharmacy72/gobak/zip"
	"os"
)

//UnPackItem Unzip the backup file to the destination folder "outDir"
func (item *BackupItem) UnPackItem(outDir string) error {
	err := zip.DoExtractFile(item.FilePath(), outDir)
	return err
}

//PackItem archives backup file,delOrgiginal - delete the original file
func (item *BackupItem) PackItem(delOrgiginal bool) (err error) {
	fileNameNoZip := item.FilePath()
	err = zip.DoZipFile(fileNameNoZip)
	if err != nil {
		return err
	}
	item.Insert = false
	item.Modified = true
	item.Status = item.Status | StatusArchived
	item.FileName = item.FileName + ".zip"
	if delOrgiginal {
		err = os.Remove(fileNameNoZip)
		if err != nil {
			return err
		}
	}
	return nil
}

//ComputeHash calculates a hash of the backup file
func (item *BackupItem) ComputeHash() error {
	hash, err := md5f.ComputeMd5String(item.FilePath())
	if err != nil {
		return err
	}
	item.Hash = hash
	return nil
}

//HashValid calculates a hash of the backup file and compares with the cache at the time of backup
func (item *BackupItem) HashValid() (bool, error) {
	return md5f.CheckMd5(item.FilePath(), item.Hash)
}

//Exists returns whether there is a backup file on disk
func (item *BackupItem) Exists() bool {
	return fileutils.Exists(item.FilePath())
}
