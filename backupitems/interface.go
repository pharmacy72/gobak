package backupitems

import "errors"

//Errors
var (
	ErrNotFound = errors.New("BackupItem not found")
)

type Repository interface {
	Close() error
	Append(item *BackupItem) error
	Update(item *BackupItem) error
	Delete(item *BackupItem) error
	Refresh(item *BackupItem) error
	All() Collection
}

type Collection interface {
	Get() ([]*BackupItem, error)
	AddFilterID(ids ...string) Collection
	ClearFilters()
	//Todo: filters
}
