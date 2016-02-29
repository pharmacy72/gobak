package backupitems

import (
	"gobak/level"
	"path/filepath"
	"testing"
)

func TestNewItem(t *testing.T) {
	s := "/path"
	r1 := New(s)
	if s != r1.basefolder {
		t.Errorf("Excepted %s, got %q", s, r1.basefolder)
	}
	r1.FileName = "file"
	r1.Level = level.NewLevel(1)
	fp := filepath.Join(s, "1", "file")
	if r1.FilePath() != fp {
		t.Errorf("Excepted %s, got %s", fp, r1.FilePath())
	}
}

func TestGetLastChainPoint(t *testing.T) {

	r1 := &BackupItem{ID: 0}
	r1_1 := &BackupItem{ID: 2, Parent: r1}
	r1_2 := &BackupItem{ID: 3, Parent: r1_1}
	r1_3 := &BackupItem{ID: 5, Parent: r1_2}

	var control = [...]*BackupItem{r1, r1_1, r1_2, r1_3}
	var result = r1_3.ChainWithAllParents()
	l := len(result)
	if l != 4 {
		t.Errorf("Expected 4, got:%d", l)
	}
	for i := range result {
		if result[i] != control[i] {
			t.Errorf("Expected:%v, got:%v", control[i], result[i])
		}
	}
}

func TestFlags(t *testing.T) {
	r1 := &BackupItem{ID: 0}
	r1.Status = 0

	if r1.IsArchived() {
		t.Error("Excepted not zipped")
	}
	r1.Status = StatusArchived
	if !r1.IsArchived() {
		t.Error("Excepted zipped")
	}

}
