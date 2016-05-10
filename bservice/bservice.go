package bservice

import (
	"errors"
	"github.com/pharmacy72/gobak/backupitems"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/dbase"
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/level"
	"log"
	"os"
	"github.com/beevik/guid"
	"path/filepath"
	"strings"
	"time"
	"fmt"
	"github.com/pharmacy72/gobak/command"
	"github.com/pharmacy72/gobak/dbfile"
	"strconv"
	"github.com/pharmacy72/gobak/errout"
)

// Errors
var (
	ErrFileDestAlreadyExists = errors.New("Distination file already exists")
	ErrFileSourceNotFound    = errors.New("Backup by filename not found")
	ErrIDSourceNotFound      = errors.New("Backup by identifier not found")
	ErrFileCorrupt           = errors.New("The backup file is corrupt")
)

func doVerbose(verbose bool, a ...interface{}) {
	if verbose {
		fmt.Printf(a[0:1][0].(string), a[1:]...)
	}
}


func wrapCmd2ErrOut(c *command.Command, reportIfError bool) *errout.ErrOut {
	return errout.New(c.Error, reportIfError, c.Stdout.Buffer, c.Stderr.Buffer)
}


//Backup make backup file for level
func Backup(verbose bool, lev level.Level, guidPrev string) (res *backupitems.BackupItem, err error) {
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

	log.Println("CMD:", args)
	cmd := command.Exec(verbose, config.Current().PathToNbackup, args[:]...)
	if cmd.Error != nil {
		//log.Print(cmd.Error)
		return nil, wrapCmd2ErrOut(cmd,true)
	}

	if err := res.ComputeHash(); err != nil {
		return res, err
	}
	res.Insert = true
	return res, nil
}

func removeUnzipFiles(verbose bool, files []string) {
	for i := 0; i < len(files); i++ {
		if e := os.Remove(files[i]); e != nil {
			log.Printf("Remove unpacked file %q error:%s\n", files[i], e)
		} else {
			doVerbose(verbose, "Removed unpacket temp file backup:%s\n", files[i])
		}
	}
}

//Restore backup into dest, optional with to checking hash
func Restore(dest string, elem *backupitems.BackupItem, hash bool, verbose bool) (res bool, err error) {
	var crapfiles []string
	defer func() {
		//remove unzipped files
		removeUnzipFiles(verbose, crapfiles)
	}()

	doVerbose(verbose, "Restore: backup(id %s) file %s\n", elem.ID, elem.FilePath())
	var restoreDest string
	if dest != "" {
		if fileutils.Exists(dest) {
			return false, ErrFileDestAlreadyExists
		}
		restoreDest = dest
	} else {
		restoreDest = fileutils.GetTempFile(config.Current().PathToBackupFolder, time.Now().Local().Format("2006-01-02_15_04")+".restore.fdb")
	}
	doVerbose(verbose, "Destination:%s\n", restoreDest)
	//Get elements from repository
	chain := elem.ChainWithAllParents()
	var backupFiles []string
	for _, n := range chain {
		var srcFile string
		if hash {
			doVerbose(verbose, "Checking hash file\n")
			if ok, err := n.HashValid(); !ok {
				if err != nil {
					return false, err
				}
				return false, ErrFileCorrupt
			}
		}
		if n.IsArchived() {
			doVerbose(verbose, "Unpack backup id(%s) file=%s", n.ID, n.FilePath())
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

	if _, err := dbfile.Restore(restoreDest, backupFiles, verbose); err != nil {
		return false, err
	}
	doVerbose(verbose, "Restored.\n")
	return true, nil
}

//RestoreFromFile Restore backup into dest from filename,optional with to checking hash
func RestoreFromFile(filename string, dest string, hash bool, verbose bool) error {
	arr, err := dbase.FetchBackupItems()
	if err != nil {
		return err
	}
	if arr == nil {
		return nil
	}
	for _, item := range arr {
		if item.FilePath() == filename {
			if _, err := Restore(dest, item, hash, verbose); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrFileSourceNotFound
}

//RestoreFromID Restore backup into dest by ID,optional with to checking hash
func RestoreFromID(id int, dest string, hash bool, verbose bool) error {
	arr, err := dbase.FetchBackupItems()
	if err != nil {
		return err
	}
	if arr == nil {
		return nil
	}
	for _, item := range arr {
		if item.ID == id {
			if _, err := Restore(dest, item, hash, verbose); err != nil {
				return err
			}
			return nil
		}
	}
	return ErrIDSourceNotFound
}
