package dotlan

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/environment"
	"github.com/GSH-LAN/Unwindia_common/src/go/config"
	"github.com/GSH-LAN/Unwindia_common/src/go/helper"
	"github.com/GSH-LAN/Unwindia_common/src/go/sql"
	"github.com/gammazero/workerpool"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

const (
	dotlanTimeout = time.Second * 10
)

type DotlanDbClient interface {
	GetTournaments(ctx context.Context, resultChan chan TournamentsResult)
	GetContestForTournament(ctx context.Context, tournament *Tournament) ([]Contest, error)
	GetTeam(ctx context.Context, teamID int) (*Team, error)
	GetTeamParticipants(ctx context.Context, teamID int) (ParticipantsTeamList, error)
	GetUser(ctx context.Context, userID int) (*User, error)
	GetUsersForTeam(ctx context.Context, teamID int) (UserList, error)
}

type DotlanDbClientImpl struct {
	ctx                 context.Context
	db                  *sqlx.DB
	env                 *environment.Environment
	lock                sync.Mutex
	workerpool          *workerpool.WorkerPool
	modelFieldCache     map[string]string
	modelFieldCacheLock sync.RWMutex
	config              config.ConfigClient
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
		ctx:                 ctx,
		db:                  db,
		env:                 env,
		workerpool:          wp,
		config:              config,
		modelFieldCache:     make(map[string]string),
		modelFieldCacheLock: sync.RWMutex{},
	}, nil
}

func (d *DotlanDbClientImpl) GetTournaments(ctx context.Context, resultChan chan TournamentsResult) {
	var tournaments []Tournament
	var err error

	ctx, cancel := context.WithTimeout(ctx, dotlanTimeout)
	defer cancel()

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

func (d *DotlanDbClientImpl) GetContestForTournament(ctx context.Context, tournament *Tournament) (contests []Contest, err error) {
	ctx, cancel := context.WithTimeout(ctx, dotlanTimeout)
	defer cancel()

	tableName, fields, err := d.getFieldsFromModelWithoutTablename(Contest{})
	if err != nil {
		return nil, err
	}
	filter := fmt.Sprintf("Tid = %v and (team_a >= 0 and team_b >= 0)", tournament.Tid)
	qry := fmt.Sprintf("select %s from %s where %s", fields, tableName, filter)
	if err := d.db.SelectContext(ctx, &contests, qry); err != nil {
		return nil, err
	}

	for i := range contests {
		contests[i].Tournament = *tournament
	}
	return
}

func (d *DotlanDbClientImpl) GetTeam(ctx context.Context, teamID int) (*Team, error) {
	ctx, cancel := context.WithTimeout(ctx, dotlanTimeout)
	defer cancel()

	tableName, fields, err := d.getFieldsFromModelWithoutTablename(Team{})
	if err != nil {
		return nil, err
	}

	qry := fmt.Sprintf("select %s from %s where tnid = ? LIMIT 1", fields, tableName)
	log.Debug().Str("query", qry).Int("teamID", teamID).Msg("prepared query for getting team")

	stmt, err := d.db.PreparexContext(ctx, qry)
	if err != nil {
		return nil, err
	}

	team := Team{}
	row := stmt.QueryRowxContext(ctx, teamID)
	if err = row.StructScan(&team); err != nil {
		return nil, err
	}

	return &team, nil
}

func (d *DotlanDbClientImpl) GetTeamParticipants(ctx context.Context, teamID int) (ParticipantsTeamList, error) {
	ctx, cancel := context.WithTimeout(ctx, dotlanTimeout)
	defer cancel()

	tableName, fields, err := d.getFieldsFromModelWithoutTablename(ParticipantsTeam{})
	if err != nil {
		return nil, err
	}
	filter := fmt.Sprintf("tnid = %d", teamID)
	qry := fmt.Sprintf("select %s from %s where %s LIMIT 1", fields, tableName, filter)

	var participants []ParticipantsTeam
	if err := d.db.SelectContext(ctx, &participants, qry); err != nil {
		return nil, err
	}

	return participants, nil
}

func (d *DotlanDbClientImpl) GetUser(ctx context.Context, userID int) (*User, error) {
	ctx, cancel := context.WithTimeout(ctx, dotlanTimeout)
	defer cancel()

	tableName, fields, err := d.getFieldsFromModelWithoutTablename(User{})
	if err != nil {
		return nil, err
	}

	qry := fmt.Sprintf("select %s from %s where id = ? LIMIT 1", fields, tableName)
	log.Debug().Str("query", qry).Int("userID", userID).Msg("prepared query for getting user")

	stmt, err := d.db.PreparexContext(ctx, qry)
	if err != nil {
		return nil, err
	}

	user := User{}
	row := stmt.QueryRowxContext(ctx, userID)
	if err = row.StructScan(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *DotlanDbClientImpl) GetUsersForTeam(ctx context.Context, teamID int) (UserList, error) {
	ctx, cancel := context.WithTimeout(ctx, dotlanTimeout*100)
	defer cancel()

	_, fields, err := d.getFieldsFromModelWithoutTablename(User{})
	if err != nil {
		return nil, err
	}

	qry := fmt.Sprintf("select %s from user u INNER JOIN t_teilnehmer_part ttp ON u.id = ttp.user_id where ttp.tnid = ?", fields)
	log.Debug().Str("query", qry).Int("teamID", teamID).Msg("prepared query for getting users for team")

	stmt, err := d.db.PreparexContext(ctx, qry)
	if err != nil {
		return nil, err
	}

	var users []User
	rows, err := stmt.QueryxContext(ctx, teamID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user User
		if err = rows.StructScan(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
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
	d.modelFieldCacheLock.RLock()
	defer d.modelFieldCacheLock.RUnlock()
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

	go func() {
		d.modelFieldCacheLock.Lock()
		defer d.modelFieldCacheLock.Unlock()
		d.modelFieldCache[tableName] = queryString
	}()

	return queryString, nil
}
