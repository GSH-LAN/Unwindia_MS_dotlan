package dotlan

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/environment"
	"github.com/GSH-LAN/Unwindia_common/src/go/config"
	"github.com/GSH-LAN/Unwindia_common/src/go/helper"
	"github.com/GSH-LAN/Unwindia_common/src/go/sql"
	"github.com/gammazero/workerpool"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type DotlanDbClient interface {
	GetTournaments(ctx context.Context, resultChan chan TournamentsResult)
	GetContestForTournament(*Tournament) ([]Contest, error)
}

type DotlanDbClientImpl struct {
	ctx             context.Context
	db              *sqlx.DB
	env             *environment.Environment
	lock            sync.Mutex
	workerpool      *workerpool.WorkerPool
	modelFieldCache map[string]string
	config          config.ConfigClient
}

func NewClient(ctx context.Context, env *environment.Environment, wp *workerpool.WorkerPool, config config.ConfigClient) (DotlanDbClient, error) {
	sqlxDsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		env.DotlanMySQLUser,
		env.DotlanMySQLPassword,
		env.DotlanMySQLHost,
		env.DotlanMySQLPort,
		env.DotlanMySQLDatabase)

	db, err := sqlx.Connect("mysql", sqlxDsn)

	if err != nil {
		return nil, err
	}

	return &DotlanDbClientImpl{
		ctx:             ctx,
		db:              db,
		env:             env,
		workerpool:      wp,
		config:          config,
		modelFieldCache: make(map[string]string),
	}, nil
}

func (d *DotlanDbClientImpl) GetTournaments(ctx context.Context, resultChan chan TournamentsResult) {
	var tournaments []Tournament
	var err error

	tableName, fieldList, err := d.getFieldsFromModelWithoutTablename(Tournament{})
	if err != nil {
		resultChan <- TournamentsResult{Result: nil, Error: err}
		return
	}

	filter := d.config.GetConfig().CmsConfig.TournamentFilter
	qry := fmt.Sprintf("select %s from %s where %s", fieldList, tableName, filter)
	log.Trace().Str("query", qry).Msgf("GetTournaments")

	if err = d.db.SelectContext(ctx, &tournaments, qry); err != nil {
		resultChan <- TournamentsResult{Result: nil, Error: err}
		return
	}

	resultChan <- TournamentsResult{Result: tournaments, Error: nil}
}

func (d *DotlanDbClientImpl) GetContestForTournament(tournament *Tournament) (contests []Contest, err error) {

	tableName, fields, err := d.getFieldsFromModelWithoutTablename(Contest{})
	if err != nil {
		return nil, err
	}
	filter := fmt.Sprintf("Tid = %v and (team_a != 0 or team_b != 0)", tournament.Tid)
	qry := fmt.Sprintf("select %s from %s where %s", fields, tableName, filter)
	if err := d.db.Select(&contests, qry); err != nil {
		return nil, err
	}

	for i := range contests {
		contests[i].Tournament = *tournament
	}
	return
}

func (d *DotlanDbClientImpl) getFieldsFromModelWithoutTablename(model sql.Table) (tableName string, fieldList string, err error) {
	if tabler, ok := model.(sql.Table); ok {
		tableName = tabler.TableName()
	} else {
		return "", "", errors.New("model is not a table")
	}
	fieldList, err = d.getFieldsFromModelWithTablename(model, tableName)
	return
}

func (d *DotlanDbClientImpl) getFieldsFromModelWithTablename(model interface{}, tableName string) (string, error) {

	if queryString, ok := d.modelFieldCache[tableName]; ok {
		log.Debug().Str("table", tableName).Str("query", queryString).Msg("Retrieved query from cache")
		return queryString, nil
	}

	var columns []string
	var queryString string

	qry := fmt.Sprintf("SELECT column_name FROM information_schema.columns WHERE table_name = '%s'", tableName)
	log.Trace().Str("query", qry).Msgf("getFieldsFromModelWithTablename")
	if err := d.db.Select(&columns, qry); err != nil {
		return "", err
	}

	reflectedModel := reflect.ValueOf(model)

	for i := 0; i < reflectedModel.NumField(); i++ {
		typeField := reflectedModel.Type().Field(i)
		if val, ok := typeField.Tag.Lookup("db"); ok {
			if helper.StringSliceContains(columns, val) {
				if len(queryString) > 0 {
					queryString = queryString + ", " + val
				} else {
					queryString = val
				}
			}
		}
	}

	d.modelFieldCache[tableName] = queryString

	return queryString, nil
}
