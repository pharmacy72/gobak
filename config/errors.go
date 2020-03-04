package config
import "errors"
var (
	ErrFolderBackupNotExists = errors.New("config: Folder for backup not found")
	ErrConfigLevel           = errors.New("config: levels not found")
	ErrNbackupNotExists      = errors.New("config: file Nbackup destination not exists")
	ErrGfixNotExists         = errors.New("config: file gfix  destination not exists")
	ErrPhysicalNotExists     = errors.New("config: Physicalpathdb destination not exists")
	ErrAliasDBNotExists      = errors.New("config: Alias DB is empty")
)
