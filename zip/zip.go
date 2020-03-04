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

type CompressApp struct {
	fileutils *fileutils.FileUtils
}

//DoZipFile pack file "filename" to the destination file "filename".zip
func (z *CompressApp) DoZipFile(filename string) error {
	//check exist file

	if z.fileutils.Exists(filename + ".zip") {
		err := fileutils.ErrFileAlreadyExists
		log.Println("err", err)
		return err
	}
	//check exist file end
	newFile, err := os.Create(filename + ".zip")
	if err != nil {
		return err
	}
	defer z.checkErrorFunc(newFile.Close)

	zipit := zip.NewWriter(newFile)
	defer z.checkErrorFunc(zipit.Close)

	zipFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer z.checkErrorFunc(zipFile.Close)

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

func (z *CompressApp) checkError(e error) {
	if e != nil {
		panic(e)
	}
}
func (z *CompressApp) checkErrorFunc(fnc tst) {
	e := fnc()
	if e != nil {
		panic(e)
	}
}
func (z *CompressApp) cloneZipItem(f *zip.File, outDir string) {
	// Create full directory path
	path := filepath.Join(outDir, f.Name)
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	z.checkError(err)
	// Clone if item is a file
	rc, err := f.Open()
	z.checkError(err)
	if !f.FileInfo().IsDir() {
		// Use os.Create() since Zip don't store file permissions.
		fileCopy, err := os.Create(path)
		z.checkError(err)
		_, err = io.Copy(fileCopy, rc)
		z.checkError(fileCopy.Close())
		z.checkError(err)
	}
	z.checkError(rc.Close())
}

//DoExtractFile Unzip the file to the destination folder
func (z *CompressApp) DoExtractFile(zipPath, outDir string) error {
	r, err := zip.OpenReader(zipPath)
	defer z.checkErrorFunc(r.Close)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		z.cloneZipItem(f, outDir)
	}
	return nil
}
