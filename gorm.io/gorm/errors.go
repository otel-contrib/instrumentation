package gorm

import (
	"gorm.io/gorm"
)

var (
	// ErrRecordNotFound record not found error
	ErrRecordNotFound = gorm.ErrRecordNotFound
	// ErrInvalidTransaction invalid transaction when you are trying to `Commit` or `Rollback`
	ErrInvalidTransaction = gorm.ErrInvalidTransaction
	// ErrNotImplemented not implemented
	ErrNotImplemented = gorm.ErrNotImplemented
	// ErrMissingWhereClause missing where clause
	ErrMissingWhereClause = gorm.ErrMissingWhereClause
	// ErrUnsupportedRelation unsupported relations
	ErrUnsupportedRelation = gorm.ErrUnsupportedRelation
	// ErrPrimaryKeyRequired primary keys required
	ErrPrimaryKeyRequired = gorm.ErrPrimaryKeyRequired
	// ErrModelValueRequired model value required
	ErrModelValueRequired = gorm.ErrModelValueRequired
	// ErrInvalidData unsupported data
	ErrInvalidData = gorm.ErrInvalidData
	// ErrUnsupportedDriver unsupported driver
	ErrUnsupportedDriver = gorm.ErrUnsupportedDriver
	// ErrRegistered registered
	ErrRegistered = gorm.ErrRegistered
	// ErrInvalidField invalid field
	ErrInvalidField = gorm.ErrInvalidField
	// ErrEmptySlice empty slice found
	ErrEmptySlice = gorm.ErrEmptySlice
	// ErrDryRunModeUnsupported dry run mode unsupported
	ErrDryRunModeUnsupported = gorm.ErrDryRunModeUnsupported
)
