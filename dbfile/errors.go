package dbfile

import "errors"

var (
	ErrDBFileNotFound       = errors.New("DBFile: file database not found")
	ErrDBFileProtected      = errors.New("DBFile: can't overwrite original database")
	ErrCheckBase            = errors.New("DBFile: check has errors")
	ErrDBFileSourceNotFound = errors.New("DBFile: for restore not found the sourses backup")
)
