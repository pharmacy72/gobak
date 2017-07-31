package level

import (
	"testing"
	"time"
)

func TestTick(t *testing.T) {
	tick := NewTick(TickYear)
	if tick != NewTick(TickYear) {
		t.Errorf("Excepted not equal  %s and %s", NewTick(TickYear), tick)
	}
}

func TestTickcasterr(t *testing.T) {
	defer func() {
		e := recover()
		if e != ErrUnknowTickValue || e == nil {
			t.Error("Excepted ErrUnknowTickValue")
		}
	}()
	NewTick("ERR")
}

func TestTickPeriodFail(t *testing.T) {
	defer func() {
		e := recover()
		if e != ErrBadTickPeriod || e == nil {
			t.Error("Excepted ErrBadTickPeriod")
		}
	}()
	NewTick("H:24")
}

func TestPriorityCurrent(t *testing.T) {
	if p := NewTick(TickYear).Priority(); p != 10 {
		t.Errorf("Excepted priority %d, got %d", 10, p)
	}
	if p := NewTick(TickMonth).Priority(); p != 9 {
		t.Errorf("Excepted priority %d, got %d", 9, p)
	}
	if p := NewTick(TickWeek).Priority(); p != 8 {
		t.Errorf("Excepted priority %d, got %d", 8, p)
	}
	if p := Tick("ZZ").Priority(); p != -1 {
		t.Errorf("Excepted priority %d, got %d", -1, p)
	}
}

func TestPriorityMore(t *testing.T) {
	const (
		eq = 0
		lt = -1
		gt = 1
		er = -2
	)
	cmpstr := map[int]string{
		eq: "Equal",
		lt: "Less",
		gt: "Great",
		er: "Error",
	}

	ty := NewTick(TickYear)
	tm := NewTick(TickMonth)
	tw := NewTick(TickWeek)

	tmp1 := NewTick(TickMonth + ":2")
	tmp2 := NewTick(TickMonth + ":3")

	test := map[[2]Tick]int{
		[2]Tick{ty, ty}: eq,
		[2]Tick{tm, tm}: eq,
		[2]Tick{tw, tw}: eq,

		[2]Tick{ty, tm}: gt,
		[2]Tick{ty, tw}: gt,
		[2]Tick{tm, tw}: gt,

		[2]Tick{tm, ty}: lt,
		[2]Tick{tw, ty}: lt,
		[2]Tick{tw, tm}: lt,

		[2]Tick{tmp1, tmp2}: eq,
		[2]Tick{ty, tmp2}:   gt,
		[2]Tick{tmp2, tw}:   gt,

		[2]Tick{Tick("ERR"), tmp1}: er,
	}

	for key, pair := range test {
		cur := Compare(key[0], key[1])
		if cur != pair {
			t.Errorf("Compare %v. Excepted %s, got %s", key, cmpstr[pair], cmpstr[cur])
		}
	}
}

func TestDecode(t *testing.T) {
	test := map[string]struct {
		M string
		P int
	}{
		"Y":    {"Y", 0},
		"H":    {"H", 0},
		"H:":   {"H", 0},
		"W:3":  {"W", 3},
		"H:2":  {"H", 2},
		"N:15": {"N", 15},
	}
	for key, pair := range test {
		tick := NewTick(key)
		m, p := tick.Decode()
		if m != pair.M || p != pair.P {
			t.Errorf("Excepted %s:%d, got %s:%d", pair.M, pair.P, m, p)
		}
	}
}

func TestInPeriod(t *testing.T) {
	test := map[struct {
		Period       int
		Max          int
		Cr           int
		On           int
		StartFromOne bool
	}]bool{
		// for minutes
		{15, 60, 2, 14, false}:  true,
		{15, 60, 14, 14, false}: true,
		{14, 60, 2, 14, false}:  false,
		{12, 60, 0, 11, false}:  true,
		{12, 60, 0, 12, false}:  false,
		{30, 60, 0, 25, false}:  true,
		{30, 60, 0, 59, false}:  false,
		{5, 60, 58, 59, false}:  true,
		// for month
		{3, 12, 1, 3, true}:   true,
		{3, 12, 9, 12, true}:  false,
		{3, 12, 10, 12, true}: true,
		{3, 12, 12, 12, true}: true,
		{3, 12, 5, 12, true}:  false,
		{6, 12, 1, 6, true}:   true,
		{6, 12, 5, 12, true}:  false,
		{6, 12, 12, 12, true}: true,
		{2, 12, 1, 2, true}:   true,
		{2, 12, 1, 3, true}:   false,
		{2, 12, 1, 6, true}:   false,
		//for days
		{10, 31, 11, 20, true}: true,
		{10, 31, 5, 20, true}:  false,
		{10, 31, 31, 31, true}: true,
		{5, 31, 1, 3, true}:    true,
		{5, 31, 10, 30, true}:  false,
		//for hours
		{3, 24, 0, 2, false}: true,

		//
		{0, 24, 1, 12, false}: true,
	}
	for key, pair := range test {
		in := InEqualPeriod(key.Period, key.Max, key.Cr, key.On, key.StartFromOne)
		if in != pair {
			t.Errorf("InEqualPeriod excepted  %t,got %t (%v)", pair, in, key)
		}
	}

}

func TestTickIsActualFail(t *testing.T) {
	defer func() {
		e := recover()
		if e != ErrUnknowTickValue || e == nil {
			t.Error("Excepted ErrUnknowTickValue")
		}
	}()

	tick := Tick("ERR")
	tick.IsActual(time.Now(), time.Now())
	t.Error("Excepted error ErrUnknowTickValue")
}
