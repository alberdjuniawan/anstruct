package core

import "errors"

var (
	ErrInvalidPrompt  = errors.New("invalid prompt")
	ErrPathTraversal  = errors.New("path traversal detected")
	ErrValidationFail = errors.New("validation failed")
	ErrParseFail      = errors.New("parse failed")
	ErrReverseFail    = errors.New("reverse failed")
	ErrHistoryEmpty   = errors.New("no history to undo")
)
