// Package gj provides convenient aliases for gojango package functions and types.
// This allows users to import as "gj" for shorter, more readable code.
//
// Usage:
//   import "github.com/epuerta9/gojango/gj"
//   
//   app := gj.New(gj.WithName("myapp"))
//   settings := gj.NewBasicSettings()
//
// This is equivalent to the full package usage:
//   import "github.com/epuerta9/gojango/pkg/gojango"
//   
//   app := gojango.New(gojango.WithName("myapp"))
//   settings := gojango.NewBasicSettings()
package gj

import (
	"github.com/epuerta9/gojango/pkg/gojango"
	"github.com/epuerta9/gojango/pkg/gojango/db"
	"github.com/epuerta9/gojango/pkg/gojango/middleware"
	"github.com/epuerta9/gojango/pkg/gojango/routing"
	"github.com/epuerta9/gojango/pkg/gojango/templates"
)

// Core Application Types and Functions
type (
	Application   = gojango.Application
	AppConfig     = gojango.AppConfig
	AppContext    = gojango.AppContext
	App           = gojango.App
	AppRoute      = gojango.Route
	Option        = gojango.Option
	AppRegistry   = gojango.Registry
	Settings      = gojango.Settings
	BasicSettings = gojango.BasicSettings
)

// Core Application Functions
var (
	New               = gojango.New
	NewBasicSettings = gojango.NewBasicSettings
	GetRegistry      = gojango.GetRegistry
	
	// Application Options
	WithName  = gojango.WithName
	WithDebug = gojango.WithDebug
	WithPort  = gojango.WithPort
)

// Database Types and Functions
type (
	Connection   = db.Connection
	Config      = db.Config
	Driver      = db.Driver
	Manager     = db.Manager
	Migrator    = db.Migrator
	Migration   = db.Migration
	EntManager  = db.EntManager
	EntClient   = db.EntClient
	SchemaAPI   = db.SchemaAPI
	Transaction = db.Transaction
	MigrateOption = db.MigrateOption
)

// Database Constants
const (
	DriverPostgres = db.DriverPostgres
	DriverSQLite   = db.DriverSQLite
	DriverMySQL    = db.DriverMySQL
)

// Database Functions
var (
	DefaultConfig    = db.DefaultConfig
	PostgresConfig   = db.PostgresConfig
	SQLiteConfig     = db.SQLiteConfig
	Open            = db.Open
	NewManager      = db.NewManager
	NewMigrator     = db.NewMigrator
	NewEntManager   = db.NewEntManager
	
	// Migration Options
	WithDisableForeignKeys = db.WithDisableForeignKeys
	WithDropColumns       = db.WithDropColumns
	WithDropIndexes       = db.WithDropIndexes
)

// Middleware Types and Functions
type (
	MiddlewareRegistry = middleware.Registry
	MiddlewareFunc     = middleware.MiddlewareFunc
)

// Middleware Functions
var (
	NewMiddlewareRegistry = middleware.NewRegistry
	GetDefaults          = middleware.GetDefaults
	GetDevelopment       = middleware.GetDevelopment
	WithoutCORS          = middleware.WithoutCORS
	Minimal              = middleware.Minimal
	
	// Individual Middleware
	RequestID       = middleware.RequestID
	Logger          = middleware.Logger
	Recovery        = middleware.Recovery
	CORS           = middleware.CORS
	SecurityHeaders = middleware.SecurityHeaders
)

// Routing Types and Functions
type (
	Router = routing.Router
	Route  = routing.Route
)

// Routing Functions
var (
	NewRouter = routing.NewRouter
)

// Template Types and Functions
type (
	Engine = templates.Engine
)

// Template Functions
var (
	NewEngine = templates.NewEngine
)