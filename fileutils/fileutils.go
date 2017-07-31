package fileutils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

//Errors when working with the files
var (
	ErrFileAlreadyExists = errors.New("file destination already exists")
	ErrFileDestIsDir     = errors.New("destination is directory")
)

// free space on directory
func FreeSpace(path string) (uint64, error) {
	fs := syscall.Statfs_t{}

	err := syscall.Statfs(path, &fs)
	if err != nil {
		return 0, err
	}

	return fs.Bavail * uint64(fs.Bsize), nil
}

//FileCopy Copy source file to dest with option overwrite
func FileCopy(source, dest string, overwrite bool) (bool, error) {
	in, err := os.Open(source)
	defer in.Close()
	if err != nil {
		return false, err
	}

	s, e := os.Stat(dest)
	if s != nil {
		if s.IsDir() {
			return false, ErrFileDestIsDir
		}
		if !overwrite {
			return false, ErrFileAlreadyExists
		}
		e = os.Remove(dest)
		if e != nil {
			return false, e
		}
	}

	out, eout := os.Create(dest)
	defer out.Close()
	if eout != nil {
		return false, eout
	}
	if _, err = io.Copy(out, in); err != nil {
		return false, err
	}
	if err = out.Sync(); err != nil {
		return false, err
	}
	return true, nil
}

//Exists check exists file for filepath
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

//MakeDirsLevels Make subdirs for each level
func MakeDirsLevels(basedir string, maxlevel int) {
	for i := 0; i <= maxlevel; i++ {
		dirlevel := filepath.Join(basedir, strconv.Itoa(i))
		if f, err := os.Stat(dirlevel); os.IsNotExist(err) || f == nil || !f.IsDir() {
			err := os.Mkdir(dirlevel, 0777)
			if err != nil {
				panic(err)
			}
		} else if err != nil {
			panic(err)
		}
	}
}

//Size returns length in bytes for regular file
func Size(path string) int64 {
	f, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return f.Size()
}

// delete file
func deleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFiles(dir string, interval int) error { // dir is the parent directory you what to search
	fmt.Println("delete folder", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		modTime := file.ModTime()

		if !modTime.After(time.Now().AddDate(0, 0, -interval)) {
			fmt.Println("filepath", filepath.Join(dir, file.Name()))
			deleteFile(filepath.Join(dir, file.Name()))

		}

	}
	return nil
}

//SizeToFredly returns length in human format for regular file
func SizeToFredly(s int64) string {
	if s < 1024 {
		return strconv.FormatInt(s, 10) + " bytes"
	}
	if s < 1024*1024 {
		return strconv.FormatFloat(float64(s)/1024, 'f', 2, 64) + "Kb"
	}
	if s < 1024*1024*1024 {
		return strconv.FormatFloat(float64(s)/1024/1024, 'f', 2, 64) + "Mb"
	}
	if s < 1024*1024*1024*1024 {
		return strconv.FormatFloat(float64(s)/1024/1024/1024, 'f', 2, 64) + "Gb"
	}
	return strconv.FormatInt(s, 10) + " bytes"
}

//GetTempFile generate a file name with a check for existing
func GetTempFile(dir, filename string) string {
	path := filepath.Join(dir, filename)
	for i := 0; ; i++ {
		_, e := os.Stat(path)
		if os.IsNotExist(e) {
			return path
		}
		path = filepath.Join(dir, strconv.Itoa(i)+filename)
	}
}
