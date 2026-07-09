package cmd

import (
	"errors"
	"strconv"
)

var ErrSilent = errors.New("silent")

type CodeError struct {
	Code int
	Err  error
}

func (e CodeError) Error() string {
	if e.Err == nil {
		return "code " + strconv.Itoa(e.Code)
	}

	return e.Err.Error()
}

func (e CodeError) Unwrap() error {
	return e.Err
}

type UsageError struct {
	Err error
}

func (e UsageError) Error() string {
	if e.Err == nil {
		return "usage"
	}

	return e.Err.Error()
}

func (e UsageError) Unwrap() error {
	return e.Err
}

type ArgsError struct {
	Err error
}

func (e ArgsError) Error() string {
	return e.Err.Error()
}

func (e ArgsError) Unwrap() error {
	return e.Err
}
