package backupitems

import (
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/md5f"
	"github.com/pharmacy72/gobak/zip"
	"os"
	
)

//UnPackItem Unzip the backup file to the destination folder "outDir"
func (r *BackupItem) UnPackItem(outDir string) error {
	err := zip.DoExtractFile(r.FilePath(), outDir)
	return err
}

//PackItem archives backup file,delOrgiginal - delete the original file
func (r *BackupItem) PackItem(delOrgiginal bool) (err error) {
	
	fileNameNoZip := r.FilePath()
	
	err = zip.DoZipFile(fileNameNoZip)
	if err != nil {
		return err
	}
	r.Insert = false
	r.Modified = true
	r.Status = r.Status | StatusArchived
	r.FileName = r.FileName + ".zip"
	if delOrgiginal {
		err = os.Remove(fileNameNoZip)
		if err != nil {
			return err
		}
	}
	return nil
}

//ComputeHash calculates a hash of the backup file
func (r *BackupItem) ComputeHash() error {
	hash, err := md5f.ComputeMd5String(r.FilePath())
	if err != nil {
		return err
	}
	r.Hash = hash
	return nil
}

//HashValid calculates a hash of the backup file and compares with the cache at the time of backup
func (r *BackupItem) HashValid() (bool, error) {
	return md5f.CheckMd5(r.FilePath(), r.Hash)
}

//Exists returns whether there is a backup file on disk
func (r *BackupItem) Exists() bool {
	return fileutils.Exists(r.FilePath())
}
