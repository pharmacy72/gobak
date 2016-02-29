package level

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

//Errors for LevelList
var (
	ErrLevelListNoSuccessively = errors.New("ListLevel: Level must be successively")
	ErrLevelAlreadyExists      = errors.New("ListLevel: level already exists")
	ErrTickPriorityBroken      = errors.New("ListLevel: broken priority tiks")
	ErrTickAlreadyExists       = errors.New("ListLevel: tick already exists")
	ErrLevelNotFound           = errors.New("ListLevel: level not found")
)

//A Item it a config item with Tick
// has parent LevelItem
type Item struct {
	Level
	Tick
	Check bool
	Prev  *Item
}

//A Levels list of the LevelItem
type Levels struct {
	items []*Item
	last  *Item
}

//NewList create list *LevelList
func NewList() *Levels {
	return &Levels{}
}

//Add new LevelItem in list
func (l *Levels) Add(level Level, tick Tick) (*Item, error) {
	if err := l.Check(level, tick); err != nil {
		return nil, err
	}
	li := &Item{
		Level: level,
		Tick:  tick,
		Check: false,
		Prev:  nil,
	}
	li.Prev = l.last

	l.last = li
	l.items = append(l.items, li)
	return li, nil
}

//First it true if LevelItem is first in a list
func (l *Levels) First() *Item {
	if len(l.items) == 0 {
		return nil
	}
	return l.items[0]
}

//Last it true if LevelItem is last in a list
func (l *Levels) Last() *Item {
	return l.last
}

//Check checks the "level" and "tick" when add
func (l *Levels) Check(level Level, tick Tick) error {
	p, e := level.Prev()
	if e == ErrLevelIsTopNotPrev && l.Exists(level) {
		return ErrLevelAlreadyExists
	}
	if e == nil && !l.Exists(p) {
		return ErrLevelListNoSuccessively
	}
	// Check priority Ticks. For priority greater level tick must be less
	for _, i := range l.items {
		cmp := Compare(tick, i.Tick)
		switch cmp {
		case 0:
			return ErrTickAlreadyExists
		case 1:
			return ErrTickPriorityBroken
		}
	}
	return nil
}

//Exists whether there is an element "item" of the current list
func (l *Levels) Exists(item Level) bool {
	if l.last == nil {
		return false
	}
	if (l.last != nil) && (item == l.last.Level) {
		return true
	}
	for n := l.last; n != nil; n = n.Prev {
		if n.Level == item {
			return true
		}
	}
	return false
}

//MaxLevel returns max level in the current list
func (l *Levels) MaxLevel() *Level {
	if l.last == nil {
		return nil
	}
	return &l.last.Level
}

//List returns slice of the list
func (l *Levels) List() []*Item {
	return l.items[:]
}

//Count returns count of the elemetns in the list
func (l *Levels) Count() int {
	return len(l.items)
}

//String it Stringer
func (l *Item) String() string {
	return fmt.Sprintf("Level:%s;Tick:%s", l.Level, l.Tick)
}

//Find returns *LevelItem by Level from the current list
func (l *Levels) Find(li Level) (*Item, error) {
	for _, i := range l.items {
		if i.Level.Equals(li) {
			return i, nil
		}
	}
	return nil, ErrLevelNotFound
}

func (l *Levels) isActualForLevelItem(item *Item, cdate time.Time, ondate time.Time) bool {
	result := item.Tick.IsActual(cdate, ondate) && (item.Prev == nil || l.isActualForLevelItem(item.Prev, cdate, ondate))
	return result
}

//IsActual Actual backup for level on date
func (l *Levels) IsActual(lvl Level, cdate time.Time, ondate time.Time) (bool, error) {
	item, err := l.Find(lvl)
	if err != nil {
		return false, err
	}
	return l.isActualForLevelItem(item, cdate, ondate), nil
}

//Schedule Get schedule for log
func (l *Levels) Schedule() string {
	result := ""
	sout := map[Tick]string{
		Tick(TickYear):   "year",
		Tick(TickMonth):  "month",
		Tick(TickWeek):   "week",
		Tick(TickDay):    "day",
		Tick(TcikHour):   "hour",
		Tick(TickMinute): "minute",
	}

	for _, item := range l.items {
		m, p := item.Tick.Decode()
		t, _ := sout[Tick(m)]
		var each string
		if p != 0 {
			each = fmt.Sprintf("each %d %ss,", p, t)
		} else {
			each = fmt.Sprintf("each %s,", t)
		}
		result += each
	}

	return strings.Trim(result, ",")
}
