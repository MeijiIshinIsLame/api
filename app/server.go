package app

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/creasty/configo"

	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/infra"
	"github.com/tadoku/api/interfaces/rdb"
	"github.com/tadoku/api/interfaces/services"
	"github.com/tadoku/api/usecases"
)

// ServerDependencies is a dependency container for the api
type ServerDependencies interface {
	AutoConfigure() error

	Init()
	Router() services.Router
	JWTGenerator() usecases.JWTGenerator
	ErrorReporter() usecases.ErrorReporter

	RDB() *infra.RDB
	SQLHandler() rdb.SQLHandler

	Repositories() *Repositories
	Interactors() *Interactors
	Services() *Services
}

// NewServerDependencies instantiates all the dependencies for the api server
func NewServerDependencies() ServerDependencies {
	return &serverDependencies{}
}

type serverDependencies struct {
	Port                 string        `envconfig:"app_port" valid:"required"`
	JWTSecret            string        `envconfig:"jwt_secret" valid:"required"`
	ErrorReporterDSN     string        `envconfig:"error_reporter_dsn"`
	SessionLength        time.Duration `envconfig:"user_session_length" valid:"required"`
	DatabaseURL          string        `envconfig:"database_url" valid:"required"`
	DatabaseMaxIdleConns int           `envconfig:"database_max_idle_conns" valid:"required"`
	DatabaseMaxOpenConns int           `envconfig:"database_max_open_conns" valid:"required"`
	CORSAllowedOrigins   []string      `envconfig:"cors_allowed_origins" valid:"required"`

	router struct {
		result services.Router
		once   sync.Once
	}

	errorReporter struct {
		result usecases.ErrorReporter
		once   sync.Once
	}

	jwtGenerator struct {
		result usecases.JWTGenerator
		once   sync.Once
	}

	rdb struct {
		result *infra.RDB
		once   sync.Once
	}

	sqlHandler struct {
		result rdb.SQLHandler
		once   sync.Once
	}

	repositories struct {
		result *Repositories
		once   sync.Once
	}

	interactors struct {
		result *Interactors
		once   sync.Once
	}

	services struct {
		result *Services
		once   sync.Once
	}
}

func (d *serverDependencies) AutoConfigure() error {
	return configo.Load(d, configo.Option{})
}

// ------------------------------
// Services
// ------------------------------

func (d *serverDependencies) Services() *Services {
	holder := &d.services
	holder.once.Do(func() {
		holder.result = NewServices(d.Interactors())
	})
	return holder.result
}

// ------------------------------
// Repositories
// ------------------------------

func (d *serverDependencies) Repositories() *Repositories {
	holder := &d.repositories
	holder.once.Do(func() {
		holder.result = NewRepositories(d.SQLHandler())
	})
	return holder.result
}

// ------------------------------
// Interactors
// ------------------------------

func (d *serverDependencies) Interactors() *Interactors {
	holder := &d.interactors
	holder.once.Do(func() {
		holder.result = NewInteractors(d.Repositories(), d.JWTGenerator(), d.SessionLength)
	})
	return holder.result
}

// ------------------------------
// Router
// ------------------------------

func (d *serverDependencies) Router() services.Router {
	holder := &d.router
	holder.once.Do(func() {
		holder.result = infra.NewRouter(d.Port, d.JWTSecret, d.CORSAllowedOrigins, d.ErrorReporter(), d.routes()...)
	})
	return holder.result
}

func (d *serverDependencies) routes() []services.Route {
	return []services.Route{
		// Service infra
		{Method: http.MethodGet, Path: "/ping", HandlerFunc: d.Services().Health.Ping},

		// Session
		{Method: http.MethodPost, Path: "/login", HandlerFunc: d.Services().Session.Login},
		{Method: http.MethodPost, Path: "/register", HandlerFunc: d.Services().Session.Register},
		{Method: http.MethodPost, Path: "/refresh", HandlerFunc: d.Services().Session.Refresh, MinRole: domain.RoleUser},

		// Users
		{Method: http.MethodPost, Path: "/users/update_password", HandlerFunc: d.Services().User.UpdatePassword, MinRole: domain.RoleUser},
		{Method: http.MethodPost, Path: "/users/profile", HandlerFunc: d.Services().User.UpdateProfile, MinRole: domain.RoleUser},

		// Contests
		{Method: http.MethodGet, Path: "/contests", HandlerFunc: d.Services().Contest.All},
		{Method: http.MethodGet, Path: "/contests/:id", HandlerFunc: d.Services().Contest.Get},
		{Method: http.MethodPost, Path: "/contests", HandlerFunc: d.Services().Contest.Create, MinRole: domain.RoleAdmin},
		{Method: http.MethodPut, Path: "/contests/:id", HandlerFunc: d.Services().Contest.Update, MinRole: domain.RoleAdmin},

		// Rankings
		{Method: http.MethodGet, Path: "/rankings/current", HandlerFunc: d.Services().Ranking.CurrentRegistration, MinRole: domain.RoleUser},
		{Method: http.MethodGet, Path: "/rankings/registration", HandlerFunc: d.Services().Ranking.RankingsForRegistration},
		{Method: http.MethodPost, Path: "/rankings", HandlerFunc: d.Services().Ranking.Create, MinRole: domain.RoleUser},
		// TODO: Rename Get to All
		{Method: http.MethodGet, Path: "/rankings", HandlerFunc: d.Services().Ranking.Get},

		// Contest logs
		{Method: http.MethodPost, Path: "/contest_logs", HandlerFunc: d.Services().ContestLog.Create, MinRole: domain.RoleUser},
		// TODO: Rename Get to All
		{Method: http.MethodGet, Path: "/contest_logs", HandlerFunc: d.Services().ContestLog.Get},
		{Method: http.MethodPut, Path: "/contest_logs/:id", HandlerFunc: d.Services().ContestLog.Update, MinRole: domain.RoleUser},
		{Method: http.MethodDelete, Path: "/contest_logs/:id", HandlerFunc: d.Services().ContestLog.Delete, MinRole: domain.RoleUser},
	}
}

func (d *serverDependencies) JWTGenerator() usecases.JWTGenerator {
	holder := &d.jwtGenerator
	holder.once.Do(func() {
		holder.result = infra.NewJWTGenerator(d.JWTSecret)
	})
	return holder.result
}

func (d *serverDependencies) ErrorReporter() usecases.ErrorReporter {
	holder := &d.errorReporter
	holder.once.Do(func() {
		var err error
		holder.result, err = infra.NewErrorReporter(d.ErrorReporterDSN)

		if err != nil {
			log.Fatalf("failed to initialize error reporter: %v\n", err)
		}

	})
	return holder.result
}

func (d *serverDependencies) Init() {
	_ = d.ErrorReporter()
}

// ------------------------------
// Relational database
// ------------------------------

func (d *serverDependencies) RDB() *infra.RDB {
	holder := &d.rdb
	holder.once.Do(func() {
		var err error
		holder.result, err = infra.NewRDB(d.DatabaseURL, d.DatabaseMaxIdleConns, d.DatabaseMaxOpenConns)

		if err != nil {
			// @TODO: we should handle errors more gracefully
			log.Fatalf("failed to initialize connection pool with database: %v\n", err)
		}
	})
	return holder.result
}

func (d *serverDependencies) SQLHandler() rdb.SQLHandler {
	holder := &d.sqlHandler
	holder.once.Do(func() {
		holder.result = infra.NewSQLHandler(d.RDB())
	})
	return holder.result
}

// RunServer starts the actual API server
func RunServer(d ServerDependencies) error {
	d.Init()

	router := d.Router()
	return router.StartListening()
}
