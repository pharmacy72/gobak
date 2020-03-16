package level

import (
	"fmt"
)

//A Level of a backup
type Level int

//NewLevel create Level
func NewLevel(level int) Level {
	return Level(level)
}

//Int cast to int
func (l Level) Int() int {
	return int(l)
}

//Prev get prev level, error if current level as root
func (l Level) Prev() (Level, error) {
	if l.IsFirst() {
		return -1, ErrLevelIsTopNotPrev
	}
	return NewLevel(l.Int() - 1), nil
}

//IsFirst returns true if current level as root
func (l Level) IsFirst() bool {
	return l.Int() == 0
}

//String it stringer
func (l Level) String() string {
	return fmt.Sprintf("%d", l.Int())
}

//Equals compare the two levels
func (l Level) Equals(level Level) bool {
	return int(l) == int(level)
}

//Next get next level for a current level
func (l Level) Next() Level {
	return NewLevel(int(l) + 1)
}
