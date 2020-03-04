package zip

import (
	"archive/zip"
	"github.com/pharmacy72/gobak/fileutils"
	"io"
	"log"
	"os"
	"path/filepath"
)

type tst func() error

//DoZipFile pack file "filename" to the destination file "filename".zip
func DoZipFile(filename string) error {
	//check exist file

	if fileutils.Exists(filename + ".zip") {
		err := fileutils.ErrFileAlreadyExists
		log.Println("err", err)
		return err
	}
	//check exist file end
	newFile, err := os.Create(filename + ".zip")
	if err != nil {
		return err
	}
	defer checkErrorFunc(newFile.Close)

	zipit := zip.NewWriter(newFile)
	defer checkErrorFunc(zipit.Close)

	zipFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer checkErrorFunc(zipFile.Close)

	// get the file information
	info, err := zipFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Method = zip.Deflate

	writer, err := zipit.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, zipFile)
	return err

}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
func checkErrorFunc(fnc tst) {
	e := fnc()
	if e != nil {
		panic(e)
	}
}
func cloneZipItem(f *zip.File, outDir string) {
	// Create full directory path
	path := filepath.Join(outDir, f.Name)
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	checkError(err)
	// Clone if item is a file
	rc, err := f.Open()
	checkError(err)
	if !f.FileInfo().IsDir() {
		// Use os.Create() since Zip don't store file permissions.
		fileCopy, err := os.Create(path)
		checkError(err)
		_, err = io.Copy(fileCopy, rc)
		checkError(fileCopy.Close())
		checkError(err)
	}
	checkError(rc.Close())
}

//DoExtractFile Unzip the file to the destination folder
func DoExtractFile(zipPath, outDir string) error {
	r, err := zip.OpenReader(zipPath)
	defer checkErrorFunc(r.Close)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		cloneZipItem(f, outDir)
	}
	return nil
}
