package backupitems

//FindByID find BackupItem in array items by ID
func FindByID(id int, items []*BackupItem) *BackupItem {
	for i := 0; i < len(items); i++ {
		if items[i].ID == id {
			return items[i]
		}
	}
	return nil
}
