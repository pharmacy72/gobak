package dbfile

import (
	"errors"
	"fmt"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/fileutils"

	"github.com/pharmacy72/gobak/command"
	"github.com/pharmacy72/gobak/errout"
	"os"
	"strings"
)

// Errors in the processing of a database file
var (
	ErrDBFileNotFound       = errors.New("DBFile: file database not found")
	ErrDBFileProtected      = errors.New("DBFile: can't overwrite original database")
	ErrCheckBase            = errors.New("DBFile: check has errors")
	ErrDBFileSourceNotFound = errors.New("DBFile: for restore not found the sourses backup")
)

//A DBFile it allows you to work with the database: check, block, restore, etc.
type DBFile struct {
	locked   bool
	verbose  bool
	Filename string
	Alias    string
	User     string
	Password string
}

//New create *DBFile
func New(filename, alias, user, password string, verbose bool) *DBFile {
	result := &DBFile{
		Filename: filename,
		Alias:    alias,
		User:     user,
		Password: password,
		verbose:  verbose,
	}
	if !result.Exists() {
		panic(ErrDBFileNotFound)
	}
	return result
}

func wrapCmd2ErrOut(c *command.Command, reportIfError bool) *errout.ErrOut {
	return errout.New(c.Error, reportIfError, c.Stdout.Buffer, c.Stderr.Buffer)
}

//Exists check exists database file
func (d *DBFile) Exists() bool {
	_, err := os.Stat(d.Filename)
	return err == nil || !os.IsNotExist(err)
}

//Lock Locked DB nbackup -L
func (d *DBFile) Lock() error {
	if d.verbose {
		fmt.Println("Lock", d.Alias)
	}
	cmd := command.Exec(d.verbose, config.Current().PathToNbackup, "-U", d.User, "-P", d.Password, "-L", d.Alias)
	if cmd.Error != nil {
		return wrapCmd2ErrOut(cmd, true)
	}
	d.locked = true
	return nil
}

//Unlock Unlocked DB nbackup -N
func (d *DBFile) Unlock(force bool) error {
	if !force && !d.locked {
		return nil
	}
	if d.verbose {
		fmt.Println("Unlock", d.Alias)
	}
	cmd := command.Exec(d.verbose, config.Current().PathToNbackup, "-U", d.User, "-P", d.Password, "-N", d.Alias)
	if cmd.Error != nil {
		return wrapCmd2ErrOut(cmd, true)
	}
	d.locked = false
	return nil
}

//IsProtected this indicates that the file DBFile production database
func IsProtected(dest string) bool {
	return strings.Compare(strings.ToUpper(config.Current().Physicalpathdb), strings.ToUpper(dest)) == 0
}

//Copy the database file to a destination folder "dest"
func (d *DBFile) Copy(dest string, overwrite bool) (*DBFile, error) {
	if IsProtected(dest) {
		return nil, ErrDBFileProtected
	}
	if d.verbose {
		fmt.Println("Copy ", d.Filename, "to", dest)
	}
	if _, err := fileutils.FileCopy(d.Filename, dest, overwrite); err != nil {
		return nil, err
	}

	if d.verbose {
		fmt.Println("Copied", d.Alias, "into ", dest)
	}
	return New(dest, dest, d.User, d.Password, d.verbose), nil
}

//Remove the database file from disk with checking production database protection
func (d *DBFile) Remove() error {
	if IsProtected(d.Filename) {
		return ErrDBFileProtected
	}
	return os.Remove(d.Filename)
}

//Fixup DB nbackup -F for
func (d *DBFile) Fixup() error {
	if d.verbose {
		fmt.Println("Fixup", d.Alias)
	}
	cmd := command.Exec(d.verbose, config.Current().PathToNbackup, "-F", d.Alias)
	if cmd.Error != nil {
		return wrapCmd2ErrOut(cmd, true)
	}
	return nil
}

//Check using gfix for full validation database
func (d *DBFile) Check() error {
	if d.verbose {
		fmt.Println("Starting check database", d.Filename)
	}
	cmd := command.Exec(d.verbose, config.Current().Pathtogfix, "-v", "-full", d.Filename, "-user", d.User, "-password", d.Password)

	if cmd.Error != nil {
	}

	outCheck := cmd.Stdout.Buffer.String()
	if cmd.Error != nil || string(outCheck) != "" {
		//outerr:=cmd.Stderr.Buffer.String()
		//outCheck +="\n"+outerr
		//smail.MailSend(string(outCheck), config.Current().AliasDb+": Check base is not correct", "", "")
		if cmd.Error != nil {
			return wrapCmd2ErrOut(cmd, true)
		}
		cmd.Error = ErrCheckBase
		return wrapCmd2ErrOut(cmd, true)
	}
	if d.verbose {
		fmt.Println("The check finished without errors ;-) Nice day!")
	}
	return nil
}

//IsLocked It indicates that the DB is in the locked(backup) mode
func (d *DBFile) IsLocked() bool {
	// must exists .delta?
	return d.locked
}

//Restore database  to a destination folder "dest" with the "files" (name of backup the files)
//with  a checking  overwrite of production database protection
func Restore(dest string, files []string, verbose bool) (*DBFile, error) {
	if IsProtected(dest) {
		return nil, ErrDBFileProtected
	}
	if len(files) == 0 {
		return nil, ErrDBFileSourceNotFound
	}
	//check exists sources
	for _, f := range files {
		if !fileutils.Exists(f) {
			return nil, ErrDBFileSourceNotFound
		}
	}

	cmd := "-R " + dest + " "
	if config.Current().Password != "" {
		cmd = "-P " + config.Current().Password + " " + cmd
	}
	if config.Current().User != "" {
		cmd = "-U " + config.Current().User + " " + cmd
	}
	c := append(strings.Fields(cmd)[:], files...)
	command := command.Exec(verbose, config.Current().PathToNbackup, c...)
	if command.Error != nil {
		return nil, wrapCmd2ErrOut(command, true)
	}
	return New(dest, dest, config.Current().User, config.Current().Password, verbose), nil
}
