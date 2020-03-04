package fileutils

import "errors"

var ErrFileAlreadyExists = errors.New("file destination already exists")
var ErrFileDestIsDir = errors.New("destination is directory")
