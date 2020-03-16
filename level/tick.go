package level

import (
	"strconv"
	"strings"
	"time"
)

//A Tick it a predetermined time period
type Tick string

//Predefined ticks
var (
	TickYear   = "Y"
	TickMonth  = "M"
	TickWeek   = "W"
	TickDay    = "D"
	TcikHour   = "H"
	TickMinute = "N"
)

var (
	//For check period and priority
	lenperiod = map[Tick]struct {
		Max      int
		FromOne  bool
		Priority int
	}{

		Tick(TickYear):   {0, false, 10}, // off
		Tick(TickMonth):  {12, true, 9},  // in year
		Tick(TickWeek):   {6, true, 8},   // in month
		Tick(TickDay):    {31, true, 7},  // in month
		Tick(TcikHour):   {24, false, 6}, // in day
		Tick(TickMinute): {60, false, 5}, // in hour
	}
)

//NewTick create Tick
func NewTick(tick string) Tick {
	m, period := Tick(tick).Decode()
	switch m {
	case TickYear, TickMonth, TickWeek, TickDay, TcikHour, TickMinute:
	default:
		panic(ErrUnknownTickValue)
	}
	if period > 0 {
		lenp, _ := lenperiod[Tick(m)]
		if lenp.Max == 0 || period >= lenp.Max {
			panic(ErrBadTickPeriod)
		}
	}
	return Tick(tick)
}

//Priority relative to other pre-defined periods
func (t Tick) Priority() int {
	m, _ := t.Decode()
	if p, ok := lenperiod[Tick(m)]; ok {
		return p.Priority
	}
	return -1
}

// Compare ticks priority
// if t1=t2 ret 0; if t1>t2 ret 1; if t1<t2 ret -1
func Compare(t1, t2 Tick) int {
	p1, p2 := t1.Priority(), t2.Priority()
	if p1 == -1 || p2 == -1 {
		return -2
	}
	if p1 > p2 {
		return 1
	} else if p1 < p2 {
		return -1
	}
	return 0
}

//Decode parses the Tick on current tick and period
// for "M:15" retruns "M",15
func (t Tick) Decode() (string, int) {
	s := strings.Split(string(t), ":")
	if len(s) <= 1 {
		return string(t), 0
	}
	i, err := strconv.Atoi(s[1])
	if err != nil {
		return s[0], 0
	}
	return s[0], i
}

func inPeriod(cur, fper, max, checkper1, checkper2, correction int) bool {
	return cur == 0 && fper >= checkper1+correction && ((fper < checkper2+correction) || (checkper2 >= max))
}
func fillPeriods(period, max int) (result []int) {
	for i := 0; i < max; i = i + period {
		result = append(result, i)
	}
	result = append(result, max)
	return result
}

//InEqualPeriod retruns the equivalent "cr" and "on"
func InEqualPeriod(period, max, cr, on int, startFromOne bool) bool {
	if period == 0 || cr > max || on > max {
		return true
	}
	correction := 0
	if startFromOne {
		correction = 1
	}
	rng := fillPeriods(period, max)
	var crperiod, onperiod int
	for i := 0; i < len(rng)-1; i++ {
		if inPeriod(crperiod, cr, max, rng[i], rng[i+1], correction) {
			crperiod = i + 1
		}
		if inPeriod(onperiod, on, max, rng[i], rng[i+1], correction) {
			onperiod = i + 1
		}
		if onperiod != 0 && crperiod != 0 {
			break
		}
	}
	return onperiod == crperiod
}

//GetWeekInMonth Get week on 1st day on month
func GetWeekInMonth(t time.Time) int {
	_, _, d := t.Date()
	_, week1st := t.AddDate(0, 0, -d+1).ISOWeek()
	_, isoweek := t.ISOWeek()
	return isoweek - week1st + 1
}

//IsActual Whether actual date "crdate" to "ondate" for the current Tick
//Attention: must check for all levels
func (t Tick) IsActual(crdate, ondate time.Time) (result bool) {
	m, p := t.Decode()
	crper, onper := 0, 0
	switch m {
	case TickYear:
		result = crdate.Year() == ondate.Year()
	case TickMonth:
		crper = int(crdate.Month())
		onper = int(ondate.Month())
		result = crdate.Month() == ondate.Month()
	case TickWeek:
		crper = GetWeekInMonth(crdate)
		onper = GetWeekInMonth(ondate)
		result = crper == onper
	case TickDay:
		crper = crdate.Day()
		onper = ondate.Day()
		result = crper == onper
	case TcikHour:
		crper = crdate.Hour()
		onper = ondate.Hour()
		result = crper == onper
	case TickMinute:
		crper = crdate.Minute()
		onper = ondate.Minute()
		result = crper == onper
	default:
		panic(ErrUnknownTickValue)
	}

	if p > 1 {
		lenp := lenperiod[Tick(m)]
		if lenp.Max > 0 {
			result = InEqualPeriod(p, lenp.Max, crper, onper, lenp.FromOne)
		}
	}

	return result
}
