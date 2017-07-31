package level

import "testing"

func TestLevel(t *testing.T) {
	l := NewLevel(1)
	if l.Int() != 1 {
		t.Errorf("Excepted: %d, got: %d", l.Int(), 1)
	}
	if l.IsFirst() {
		t.Error("Excepted level are not top")
	}
	if l.String() != "1" {
		t.Errorf("Excepted 1 got %s", l)
	}
	if !l.Equals(NewLevel(1)) {
		t.Errorf("Excepted equals %s and %s", NewLevel(1), l)
	}
	if l.Equals(NewLevel(0)) {
		t.Errorf("Excepted not equals %s and %s", NewLevel(0), l)
	}
	prev, err := l.Prev()
	if err != nil {
		t.Error(err)
	}
	if prev != NewLevel(0) {
		t.Errorf("Excepted: %s, got: %s", NewLevel(0), prev)
	}

	if !prev.IsFirst() {
		t.Error("Excepted level must be top")
	}
	prev, err = prev.Prev()
	if err != ErrLevelIsTopNotPrev {
		t.Error("Exception error=ErrLevelIsTopNotPrev")
	}
}

func TestLevelNext(t *testing.T) {
	l := NewLevel(0)
	if l.Next().Equals(l) {
		t.Errorf("Level Next %s excepted not equal %s", l.Next(), l)
	}
	if !l.Next().Equals(NewLevel(1)) {
		t.Errorf("Level Next %s excepted equal %s", l.Next(), NewLevel(1))
	}

	p, _ := l.Next().Prev()
	if !p.Equals(l) {
		t.Errorf("Level Next.Prev %s excepted equal %s", p, l)
	}
}
