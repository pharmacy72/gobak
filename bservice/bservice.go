package bservice

/*
do
  Backup
  Restore
  RestoreFromFile
  RestoreFromID
*/
import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/guid"
	"github.com/pharmacy72/gobak/backupitems"
	"github.com/pharmacy72/gobak/command"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/dbfile"
	"github.com/pharmacy72/gobak/errout"
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/level"
)

type Bservice struct {
	log       *zap.Logger
	fileutils *fileutils.FileUtils
	dbfile    *dbfile.DBFile
}

func New(log *zap.Logger, fileutils *fileutils.FileUtils, dbfile *dbfile.DBFile) *Bservice {
	return &Bservice{
		log:       log,
		fileutils: fileutils,
		dbfile:    dbfile,
	}
}

func (b *Bservice) doVerbose(verbose bool, a ...interface{}) {
	if verbose {
		fmt.Printf(a[0:1][0].(string), a[1:]...)
	}
}

func (b *Bservice) wrapCmd2ErrOut(c *command.Command, reportIfError bool) *errout.ErrOut {
	return errout.New(c.Error, reportIfError, c.Stdout.Buffer, c.Stderr.Buffer)
}

//Backup make backup file for level
func (b *Bservice) Backup(verbose bool, lev level.Level, guidPrev string) (res *backupitems.BackupItem, err error) {

	res = backupitems.New(config.Current().PathToBackupFolder)
	res.GUID = guid.New().String()
	res.GUIDParent = guidPrev
	res.Date = time.Now().Local()
	res.FileName = res.Date.Format("2006-01-02_15_04") + "level_" + strconv.Itoa(lev.Int()) + ".nbk"
	res.Level = lev

	var args []string
	if config.Current().DirectIO {
		args = append(args, "-D", "ON")
	}
	args = append(args, "-U", config.Current().User, "-P",
		config.Current().Password, "-B", strconv.Itoa(lev.Int()), config.Current().Physicalpathdb, res.FilePath())

	b.log.Info(fmt.Sprintf("CMD:", args))
	cmd := command.Exec(verbose, config.Current().PathToNbackup, args[:]...)
	if cmd.Error != nil {
		return nil, b.wrapCmd2ErrOut(cmd, true)
	}

	if err := res.ComputeHash(); err != nil {
		return res, err
	}
	res.Insert = true
	return res, nil
}

func (b *Bservice) removeUnzipFiles(verbose bool, files []string) {
	for i := 0; i < len(files); i++ {
		if e := os.Remove(files[i]); e != nil {
			b.log.Info(fmt.Sprintf("Remove unpacked file %q error:%s\n", files[i], e))
		} else {
			b.doVerbose(verbose, "Removed unpacket temp file backup:%s\n", files[i])
		}
	}
}

//Restore backup into dest, optional with to checking hash
func (b *Bservice) Restore(dest string, elem *backupitems.BackupItem, hash bool, verbose bool) (res bool, err error) {
	var crapfiles []string
	defer func() {
		//remove unzipped files
		b.removeUnzipFiles(verbose, crapfiles)
	}()

	b.doVerbose(verbose, "Restore: backup(id %s) file %s\n", elem.ID, elem.FilePath())
	var restoreDest string
	if dest != "" {
		if b.fileutils.Exists(dest) {
			return false, ErrFileDestAlreadyExists
		}
		restoreDest = dest
	} else {
		restoreDest = b.fileutils.GetTempFile(config.Current().PathToBackupFolder, time.Now().Local().Format("2006-01-02_15_04")+".restore.fdb")
	}
	b.doVerbose(verbose, "Destination:%s\n", restoreDest)
	//Get elements from repository
	chain := elem.ChainWithAllParents()
	var backupFiles []string
	for _, n := range chain {
		var srcFile string
		if hash {
			b.doVerbose(verbose, "Checking hash file\n")
			if ok, err := n.HashValid(); !ok {
				if err != nil {
					return false, err
				}
				return false, ErrFileCorrupt
			}
		}
		if n.IsArchived() {
			b.doVerbose(verbose, "Unpack backup id(%s) file=%s", n.ID, n.FilePath())
			if err := n.UnPackItem(config.Current().PathToBackupFolder); err != nil {
				return false, err
			}
			srcFile = filepath.Join(config.Current().PathToBackupFolder, strings.Trim(filepath.Base(n.FileName), ".zip"))
			crapfiles = append(crapfiles, srcFile)
		} else {
			srcFile = n.FilePath()
		}
		backupFiles = append(backupFiles, srcFile)
	}

	if _, err := b.dbfile.Restore(restoreDest, backupFiles, verbose); err != nil {
		return false, err
	}
	b.doVerbose(verbose, "Restored.\n")
	return true, nil
}

//RestoreFromFile Restore backup into dest from filename,optional with to checking hash
func (b *Bservice) RestoreFromFile(filename string, dest string, hash bool, verbose bool) error {
	repo := backupitems.GetRepository()
	defer repo.Close()
	backups := repo.All()
	arr, err := backups.Get()
	if err != nil {
		return err
	}
	if arr == nil {
		return nil
	}
	for _, item := range arr {
		if item.FilePath() == filename {
			if _, err := b.Restore(dest, item, hash, verbose); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrFileSourceNotFound
}

//RestoreFromID Restore backup into dest by ID,optional with to checking hash
func (b *Bservice) RestoreFromID(id int, dest string, hash bool, verbose bool) error {
	repo := backupitems.GetRepository()
	defer repo.Close()
	backups := repo.All()
	backups.AddFilterID(strconv.Itoa(id))
	arr, err := backups.Get()
	if err != nil {
		return err
	}
	if arr == nil {
		return ErrIDSourceNotFound
	}
	for _, item := range arr {
		if _, err := b.Restore(dest, item, hash, verbose); err != nil {
			return err
		}
		return nil
	}
	return ErrIDSourceNotFound
}
