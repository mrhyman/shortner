package model

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidLinkID         = errors.New("provided id is not valid")
	ErrInvalidURL            = errors.New("provided url is not valid string")
	ErrInvalidName           = errors.New("provided name is not valid string")
	ErrNotFound              = errors.New("url not found")
	ErrLinkIsGone            = errors.New("link is gone")
	ErrShortLinkGeneration   = errors.New("short link generation error")
	ErrShortenError          = errors.New("shorten error")
	ErrInvalidRequestParams  = errors.New("invalid request params")
	ErrInvalidRequestHeaders = errors.New("ivalid request headers")
	ErrWentWrong             = errors.New("something went wrong")
	ErrEnvParsing            = errors.New("parsing .env error")
	ErrEmptyEnv              = errors.New("empty .env variable error")
	ErrLoggerSetup           = errors.New("logger setup error")
	ErrResponseEncoding      = errors.New("response encoding error")
	ErrCompressReading       = errors.New("compress reading error")
	ErrStorageInit           = errors.New("storage init error")
	ErrStorageUnavailable    = errors.New("storage unavailable")
	ErrUnknownUser           = errors.New("userID is not provided")
	ErrCookieDecoding        = errors.New("can't decode cookie")
	ErrCookieEncoding        = errors.New("can't encode cookie")
)

// generate:reset
type AlreadyExistsError struct {
	ShortURL string
	Err      error
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("lint with short url already exists: %s", e.ShortURL)
}

func (e *AlreadyExistsError) Unwrap() error {
	return e.Err
}

func NewAlreadyExistsError(shortURL string, err error) error {
	return &AlreadyExistsError{
		ShortURL: shortURL,
		Err:      err,
	}
}
