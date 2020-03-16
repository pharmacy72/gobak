package level

import "errors"

var (
	ErrLevelIsTopNotPrev       = errors.New("level: level is top and not exists prev")
	ErrLevelListNoSuccessively = errors.New("ListLevel: Level must be successively")
	ErrLevelAlreadyExists      = errors.New("ListLevel: level already exists")
	ErrTickPriorityBroken      = errors.New("ListLevel: broken priority tiks")
	ErrTickAlreadyExists       = errors.New("ListLevel: tick already exists")
	ErrLevelNotFound           = errors.New("ListLevel: level not found")
	ErrUnknownTickValue        = errors.New("unknown tick")
	ErrBadTickPeriod           = errors.New("bad period tick")
)
