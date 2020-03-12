package bservice

import "errors"

var (
	ErrFileDestAlreadyExists = errors.New("destination file already exists")
	ErrFileSourceNotFound    = errors.New("backup by filename not found")
	ErrIDSourceNotFound      = errors.New("backup by identifier not found")
	ErrFileCorrupt           = errors.New("the backup file is corrupt")
)
