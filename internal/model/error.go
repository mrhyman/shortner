package model

import (
	"errors"
)

var (
	ErrInvalidLinkId         = errors.New("provided id is not valid UUID")
	ErrInvalidUrl            = errors.New("provided url is not valid string")
	ErrInvalidName           = errors.New("provided name is not valid string")
	ErrNotFound              = errors.New("url not found")
	ErrShortLinkGeneration   = errors.New("short link generation error")
	ErrInvalidRequestParams  = errors.New("ivalid request params")
	ErrInvalidRequestHeaders = errors.New("ivalid request headers")
)
