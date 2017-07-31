package level

import (
	"testing"
	"time"
)

func TestLevelList(t *testing.T) {
	list := NewList()

	if list.MaxLevel() != nil {
		t.Errorf("Excepted maxlevel nil")
	}
	if list.First() != nil {
		t.Errorf("Excepted first must be nil")
	}

	if list.Last() != nil {
		t.Errorf("Excepted last must be nil")
	}

	if list.Count() != 0 {
		t.Errorf("Excepted Count must be zero")
	}

	li, err := list.Add(NewLevel(0), NewTick(TickYear))
	if err != nil {
		t.Error(err)
	}
	if li == nil {
		t.Error("Excepted Add retrun not nil")
	}
	if list.First() != li {
		t.Errorf("Excepted %v must be first in list", li)
	}
	if list.First() != li {
		t.Errorf("Excepted %v must be last in list", li)
	}

	if list.MaxLevel() != &li.Level {
		t.Errorf("Excepted maxlevel %v, got:%v", li.Level, list.MaxLevel())
	}

	li2, e := list.Add(NewLevel(1), NewTick(TickMonth))
	if e != nil {
		t.Error(e)
	}
	if list.First() == li2 {
		t.Errorf("Excepted %v can't be first in list", li2)
	}
	if list.last != li2 {
		t.Errorf("Excepted %v must be last in list", li2)
	}

	if li2.Prev != li {
		t.Errorf("Excepted Prev %v, got %v", li, li2.Prev)
	}

	li3, _ := list.Add(NewLevel(2), NewTick(TickWeek))

	if list.Count() != 3 {
		t.Errorf("Excepted Count must be 3,got: %d ", list.Count())
	}

	if list.MaxLevel() != &li3.Level {
		t.Errorf("Excepted maxlevel %v, got:%v", li3.Level, list.MaxLevel())
	}

	cp := list.List()

	if len(cp) != 3 {
		t.Errorf("Excepted len List() %d, got %d", 3, len(cp))
	}
	if cp[0] != li || cp[1] != li2 || cp[2] != li3 {
		t.Errorf("Excepted list() %v, got %v ", []*Item{li, li2, li3}, cp)
	}

	f, errfind := list.Find(cp[1].Level)
	if errfind != nil {
		t.Errorf("Excepted find by level %v, but error %v", cp[1], errfind)
	}
	if f != cp[1] {
		t.Errorf("Excepted find by level %v#, got %#v", cp[1], f)
	}

	f, errfind = list.Find(NewLevel(4))
	if f != nil {
		t.Errorf("Excepted result find by level nil,got %#v", f)
	}
	if errfind != ErrLevelNotFound {
		t.Errorf("Excepted find by level error ErrLevelNotFound, got %v", errfind)
	}

}

func TestLevelOrder(t *testing.T) {
	list := NewList()
	_, err := list.Add(NewLevel(1), NewTick(TickYear))
	if err != ErrLevelListNoSuccessively {
		t.Error("Excepted must be error ErrLevelListNoSuccessively")
	}

	l1, e := list.Add(NewLevel(0), NewTick(TickYear))
	if e != nil {
		t.Error(e)
	}
	if !list.Exists(l1.Level) {
		t.Errorf("Excepted exists %s", l1)
	}

	if _, e := list.Add(NewLevel(0), NewTick(TickYear)); e != ErrLevelAlreadyExists {
		t.Error("Excepted must be error ErrLevelAlreadyExists")
	}

	l2, e2 := list.Add(NewLevel(1), NewTick(TickMonth))
	if e2 != nil {
		t.Error(e2)
	}

	_, e3 := list.Add(NewLevel(3), NewTick(TickWeek))
	if e3 != ErrLevelListNoSuccessively {
		t.Error("Excepted must be error ErrLevelListNoSuccessively")
	}
	if !list.Exists(l2.Level) {
		t.Errorf("Excepted exists %s", l2)
	}
	if list.Exists(NewLevel(4)) {
		t.Errorf("Excepted not exists level %s", NewLevel(4))
	}

	if !list.Exists(l1.Level) {
		t.Errorf("Excepted exists %s", l1)
	}

}

func TestTickOrder(t *testing.T) {

	list := NewList()
	_, e := list.Add(NewLevel(0), NewTick(TickMonth))
	if e != nil {
		t.Error(e)
	}

	if _, e := list.Add(NewLevel(1), NewTick(TickYear)); e != ErrTickPriorityBroken {
		t.Error("Excepted: must be error ErrTickPriority")
	}

	if _, e := list.Add(NewLevel(1), NewTick(TickWeek)); e != nil {
		t.Error(e)
	}

	if _, e := list.Add(NewLevel(1), NewTick(TickWeek)); e != ErrTickAlreadyExists {
		t.Error("Excepted ErrTickAlreadyExists, got ", e)
	}

}

func TestActualPeriodFail(t *testing.T) {
	list := NewList()
	if _, err := list.Add(NewLevel(0), NewTick(TickYear)); err != nil {
		t.Error(err)
	}
	_, err := list.IsActual(NewLevel(1), time.Now(), time.Now())
	if err != ErrLevelNotFound {
		t.Errorf("Excepted error ErrLevelNotFound, got %v", err)
	}

}

func TestActualPeriodDayOnMonth(t *testing.T) {
	list := NewList()
	level0 := NewLevel(0)
	_, err := list.Add(level0, NewTick(TickDay+":10"))
	if err != nil {
		t.Error(err)
	}
	//Days
	check := map[struct {
		level  Level
		crdate time.Time
		ondate time.Time
	}]bool{
		{
			level0,
			time.Date(2015, 2, 2, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 9, 00, 13, 0, 0, time.Local),
		}: true,
		{
			level0,
			time.Date(2015, 2, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 10, 00, 13, 0, 0, time.Local),
		}: true,
		{
			level0,
			time.Date(2015, 2, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 11, 00, 13, 0, 0, time.Local),
		}: false,
		{
			level0,
			time.Date(2015, 2, 23, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 28, 00, 13, 0, 0, time.Local),
		}: true,
	}
	checkRange(t, list, check)
}

func TestActualPeriodHoursMinutes(t *testing.T) {
	list := NewList()
	level0 := NewLevel(0)
	level1 := NewLevel(1)
	level2 := NewLevel(2)
	_, err := list.Add(level0, NewTick(TickDay))
	if err != nil {
		t.Error(err)
	}
	_, err = list.Add(level1, NewTick(TcikHour+":3"))
	if err != nil {
		t.Error(err)
	}
	_, err = list.Add(level2, NewTick(TickMinute+":10"))
	if err != nil {
		t.Error(err)
	}

	check := map[struct {
		level  Level
		crdate time.Time
		ondate time.Time
	}]bool{
		//Hours
		{
			level1,
			time.Date(2015, 2, 2, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 2, 00, 13, 0, 0, time.Local),
		}: true,
		{
			level1,
			time.Date(2015, 2, 2, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 2, 02, 13, 0, 0, time.Local),
		}: true,
		{
			level1,
			time.Date(2015, 2, 2, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 3, 00, 13, 0, 0, time.Local),
		}: false,
		// Minutes

		{
			level2,
			time.Date(2015, 2, 2, 00, 0, 0, 0, time.Local),
			time.Date(2015, 2, 2, 00, 1, 0, 0, time.Local),
		}: true,
		{
			level2,
			time.Date(2015, 2, 2, 00, 0, 0, 0, time.Local),
			time.Date(2015, 2, 2, 00, 11, 0, 0, time.Local),
		}: false,
		{
			level2,
			time.Date(2015, 2, 2, 00, 50, 0, 0, time.Local),
			time.Date(2015, 2, 2, 00, 59, 0, 0, time.Local),
		}: true,
	}
	checkRange(t, list, check)
}

func checkRange(t *testing.T,
	list *Levels,
	check map[struct {
		level  Level
		crdate time.Time
		ondate time.Time
	}]bool) {
	for key, pair := range check {
		ok, err := list.IsActual(key.level, key.crdate, key.ondate)
		if err != nil {
			t.Error(err)
		}
		if ok != pair {
			t.Log(key.level, " - ", key.crdate, key.ondate)
			t.Errorf("Excepted IsActual %t,  got %t, date=%v ondate=%v",
				pair, !pair, key.crdate, key.ondate)
		}
	}

}

func TestActualPeriodMonthAndWeeks(t *testing.T) {
	list := NewList()
	level0 := NewLevel(0)
	level1 := NewLevel(1)
	level2 := NewLevel(2)

	var err error
	_, err = list.Add(level0, NewTick(TickMonth+":3"))
	if err != nil {
		t.Error(err)
	}
	_, err = list.Add(level1, NewTick(TickWeek+":2"))
	if err != nil {
		t.Error(err)
	}

	_, err = list.Add(level2, NewTick(TickDay+":10"))
	if err != nil {
		t.Error(err)
	}
	check := map[struct {
		level  Level
		crdate time.Time
		ondate time.Time
	}]bool{
		// month
		{
			level0,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 3, 2, 00, 13, 0, 0, time.Local),
		}: true,
		{
			level0,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 5, 2, 00, 13, 0, 0, time.Local),
		}: false,
		{
			level0,
			time.Date(2015, 5, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 8, 2, 00, 13, 0, 0, time.Local),
		}: false,
		{
			level0,
			time.Date(2015, 5, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 9, 2, 00, 13, 0, 0, time.Local),
		}: false,
		{
			level0,
			time.Date(2015, 9, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 12, 2, 00, 13, 0, 0, time.Local),
		}: false,
		{
			level0,
			time.Date(2015, 12, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 12, 2, 00, 13, 0, 0, time.Local),
		}: true,
		//Weeks

		{
			level1,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 2, 00, 13, 0, 0, time.Local),
		}: true,
		{
			level1,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 5, 00, 13, 0, 0, time.Local),
		}: true,
		{
			level1,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 12, 00, 13, 0, 0, time.Local),
		}: false,

		{
			level1,
			time.Date(2015, 2, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 2, 00, 13, 0, 0, time.Local),
		}: true,

		{
			level1,
			time.Date(2015, 2, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 9, 00, 13, 0, 0, time.Local),
		}: false,
	}
	checkRange(t, list, check)
}

func TestActualYMWDHN(t *testing.T) {
	list := NewList()
	level0 := NewLevel(0)
	level1 := NewLevel(1)
	level2 := NewLevel(2)
	level3 := NewLevel(3)
	level4 := NewLevel(4)
	level5 := NewLevel(5)
	if _, e := list.Add(level0, NewTick(TickYear)); e != nil {
		t.Error(e)
	}
	if _, e := list.Add(level1, NewTick(TickMonth)); e != nil {
		t.Error(e)
	}
	if _, e := list.Add(level2, NewTick(TickWeek)); e != nil {
		t.Error(e)
	}
	if _, e := list.Add(level3, NewTick(TickDay)); e != nil {
		t.Error(e)
	}
	if _, e := list.Add(level4, NewTick(TcikHour)); e != nil {
		t.Error(e)
	}
	if _, e := list.Add(level5, NewTick(TickMinute)); e != nil {
		t.Error(e)
	}
	// Y M W D H
	check := map[struct {
		level  Level
		crdate time.Time
		ondate time.Time
	}]bool{
		// Years
		{
			level0,
			time.Date(2014, 1, 15, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 15, 00, 13, 0, 0, time.Local),
		}: false,
		{
			level0,
			time.Date(2015, 2, 16, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 15, 00, 13, 0, 0, time.Local),
		}: true,

		//Months
		{
			level1,
			time.Date(2014, 1, 15, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 15, 00, 13, 0, 0, time.Local),
		}: false,
		{
			level1,
			time.Date(2015, 1, 15, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 15, 01, 13, 0, 0, time.Local),
		}: false,
		{
			level1,
			time.Date(2015, 1, 15, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 15, 01, 13, 0, 0, time.Local),
		}: true,

		// Weeks
		{
			level2,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 2, 1, 01, 13, 0, 0, time.Local),
		}: false,
		{
			level2,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 5, 00, 00, 1, 0, time.Local),
		}: false,
		{
			level2,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 4, 00, 00, 1, 0, time.Local),
		}: true,
		// Days
		{
			level3,
			time.Date(2015, 1, 1, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 4, 00, 00, 1, 0, time.Local),
		}: false,
		{
			level3,
			time.Date(2015, 1, 2, 00, 13, 0, 0, time.Local),
			time.Date(2015, 1, 2, 00, 00, 1, 0, time.Local),
		}: true,
		{
			level3,
			time.Date(2015, 1, 1, 00, 00, 0, 0, time.Local),
			time.Date(2015, 2, 1, 00, 00, 0, 0, time.Local),
		}: false,
		// Hours
		{
			level4,
			time.Date(2015, 1, 2, 00, 13, 0, 10, time.Local),
			time.Date(2015, 1, 2, 00, 00, 1, 0, time.Local),
		}: true,
		{
			level4,
			time.Date(2015, 1, 2, 00, 13, 0, 10, time.Local),
			time.Date(2015, 1, 2, 01, 00, 1, 0, time.Local),
		}: false,
		{
			level4,
			time.Date(2015, 1, 2, 00, 13, 0, 10, time.Local),
			time.Date(2015, 1, 3, 00, 00, 1, 0, time.Local),
		}: false,
		// Minute
		{
			level5,
			time.Date(2015, 1, 2, 00, 01, 0, 0, time.Local),
			time.Date(2015, 1, 2, 00, 00, 0, 0, time.Local),
		}: false,
		{
			level5,
			time.Date(2015, 1, 2, 00, 00, 1, 0, time.Local),
			time.Date(2015, 1, 2, 00, 00, 0, 0, time.Local),
		}: true,
		{
			level5,
			time.Date(2015, 1, 2, 00, 00, 0, 0, time.Local),
			time.Date(2015, 1, 3, 00, 00, 0, 0, time.Local),
		}: false,
	}

	checkRange(t, list, check)
}

func TestLevelItemString(t *testing.T) {
	list := NewList()

	sch := list.Schedule()
	if sch != "" {
		t.Errorf("Excepted schedule emtpy, got %s", sch)
	}

	li, _ := list.Add(NewLevel(0), NewTick(TickYear))
	s := "Level:0;Tick:" + TickYear
	if li.String() != s {
		t.Errorf("Excepted Levelitem string %s,got %s", s, li.String())
	}

	_, err := list.Add(NewLevel(1), NewTick(TickMonth+":2"))
	if err != nil {
		t.Error(err)
	}

	sch = list.Schedule()
	if sch == "" {
		t.Errorf("Excepted schedule not emtpy")
	}

}
