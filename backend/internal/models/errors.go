package models

import "errors"

// ErrAlreadyFinished is returned when a terminal generation is asked to change
// state. Defined here so callers can react without importing gorm.
var ErrAlreadyFinished = errors.New("generation has already finished")
