package backupitems

import (
	"os"
)

//UnPackItem Unzip the backup file to the destination folder "outDir"
func (r *BackupItem) UnPackItem(outDir string) error {
	err := r.zip.DoExtractFile(r.FilePath(), outDir)
	return err
}

//PackItem archives backup file,delOrgiginal - delete the original file
func (r *BackupItem) PackItem(delOriginal bool) (err error) {

	fileNameNoZip := r.FilePath()

	err = r.zip.DoZipFile(fileNameNoZip)
	if err != nil {
		return err
	}
	r.Insert = false
	r.Modified = true
	r.Status = r.Status | StatusArchived
	r.FileName = r.FileName + ".zip"
	if delOriginal {
		err = os.Remove(fileNameNoZip)
		if err != nil {
			return err
		}
	}
	return nil
}

//ComputeHash calculates a hash of the backup file
func (r *BackupItem) ComputeHash() error {
	hash, err := r.md5App.ComputeMd5String(r.FilePath())
	if err != nil {
		return err
	}
	r.Hash = hash
	return nil
}

//HashValid calculates a hash of the backup file and compares with the cache at the time of backup
func (r *BackupItem) HashValid() (bool, error) {
	return r.md5App.CheckMd5(r.FilePath(), r.Hash)
}

//Exists returns whether there is a backup file on disk
func (r *BackupItem) Exists() bool {
	return r.fileutils.Exists(r.FilePath())
}
