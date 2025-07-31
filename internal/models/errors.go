package models

import "errors"

// Common errors used across models
var (
	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrWeakPassword       = errors.New("password does not meet requirements")

	// Tenant errors
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantInactive      = errors.New("tenant is inactive")
	ErrTenantAlreadyExists = errors.New("tenant already exists")

	// Video errors
	ErrVideoNotFound      = errors.New("video not found")
	ErrVideoAlreadyExists = errors.New("video already exists")
	ErrInvalidVideoFormat = errors.New("invalid video format")
	ErrVideoProcessing    = errors.New("video is currently being processed")

	// Publication errors
	ErrPublicationNotFound = errors.New("publication job not found")
	ErrPublicationFailed   = errors.New("publication failed")
	ErrInvalidPlatform     = errors.New("invalid platform")

	// General errors
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternalError = errors.New("internal server error")
	ErrNotFound      = errors.New("resource not found")
	ErrConflict      = errors.New("resource conflict")
)
