package worker

import (
	"context"
	"fmt"
	"github.com/GSH-LAN/Unwindia_common/src/go/matchservice"
	"github.com/GSH-LAN/Unwindia_common/src/go/messagebroker"
	"github.com/GSH-LAN/Unwindia_common/src/go/workitemLock"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/hashicorp/go-multierror"

	"github.com/segmentio/ksuid"
	"strconv"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/database"
	"github.com/GSH-LAN/Unwindia_MS_dotlan/src/dotlan"
	"github.com/GSH-LAN/Unwindia_common/src/go/config"
	"github.com/gammazero/workerpool"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

type Worker interface {
	Start(ticker *time.Ticker) error
}

type WorkerImpl struct {
	ctx              context.Context
	workerpool       *workerpool.WorkerPool
	config           config.ConfigClient
	dotlanClient     dotlan.DotlanDbClient
	dbClient         database.DatabaseClient
	messagePublisher message.Publisher
	semaphore        *semaphore.Weighted
	dotlanLock       workitemLock.WorkItemLock
}

func NewWorker(ctx context.Context, workerpool *workerpool.WorkerPool, config config.ConfigClient, dotlanClient dotlan.DotlanDbClient, dbClient database.DatabaseClient, publisher message.Publisher, workitemLock workitemLock.WorkItemLock) Worker {
	w := WorkerImpl{
		ctx:              ctx,
		workerpool:       workerpool,
		config:           config,
		dotlanClient:     dotlanClient,
		dbClient:         dbClient,
		messagePublisher: publisher,
		semaphore:        semaphore.NewWeighted(int64(1)),
		dotlanLock:       workitemLock,
	}
	return &w
}

func (w *WorkerImpl) Start(ticker *time.Ticker) error {
	go w.process()
	for {
		select {
		case <-ticker.C:
			go w.process()
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
}

func (w *WorkerImpl) getMatchData() (statusList database.DotlanStatusList, dotlanTournamentList dotlan.TournamentList, errGroup error) {
	var chanDotlanStates = make(chan database.Result)
	var chanTournaments = make(chan dotlan.TournamentsResult)

	w.workerpool.Submit(func() { w.dotlanClient.GetTournaments(w.ctx, chanTournaments) })
	w.workerpool.Submit(func() { w.dbClient.List(w.ctx, nil, chanDotlanStates) })

	for i := 0; i < 2; i++ {
		select {
		case tournamentsResult := <-chanTournaments:
			if tournamentsResult.Error != nil {
				errGroup = multierror.Append(errGroup, fmt.Errorf("tournamentsRetrievalError: %w", tournamentsResult.Error))
			}
			dotlanTournamentList = tournamentsResult.Result
		case dotlanStatesResult := <-chanDotlanStates:
			if dotlanStatesResult.Error != nil {
				errGroup = multierror.Append(errGroup, fmt.Errorf("statesRetrievalError: %w", dotlanStatesResult.Error))
			}
			statusList = dotlanStatesResult.Result
		}
	}

	return
}

// process is some ugly shit function that encapsulates the most business logic regarding detection of dotlan matches
// this whole thing is a mess and needs to be refactored and splitted up into multiple functions
// TODO: refactor this function
func (w *WorkerImpl) process() {
	var tournaments []dotlan.Tournament

	if !w.semaphore.TryAcquire(1) {
		log.Warn().Msg("Skip processing, semaphore already acquired")
		return
	}
	defer w.semaphore.Release(1)

	log.Debug().Msg("Start processing")

	dotlanStates, tournaments, err := w.getMatchData()
	if err != nil {
		log.Error().Err(err).Msg("Error getting match data")
	}

	log.Debug().Int("tournamentAmount", len(tournaments)).Int("dotlanAmount", len(dotlanStates)).Msg("Finished fetching stuff")
	log.Trace().Interface("tournaments", tournaments).Msg("Finished fetching stuff")

	// 2.1 Check if we have tournaments that are closed now, but we have in DB for management, if so, emit event (and maybe generate notification?)
	// retrieve missing tournaments and compare state, maybe tournament is just paused?

	// 2.2 Get all belongig contests

	// 3. Check if contests has changed and need new event published

	// now go the other way round, to check for contests that we don't longer manage...

	for i, _ := range tournaments {
		tournament := tournaments[i]
		w.workerpool.Submit(func() { w.processTournament(tournament, &dotlanStates) })
	}
}

func (w *WorkerImpl) lockContest(id uint) bool {
	if err := w.dotlanLock.Lock(w.ctx, strconv.Itoa(int(id)), nil); err != nil {
		log.Warn().Uint("contest", id).Err(err).Msg("Error locking contest")
		return false
	}
	log.Trace().Uint("contest", id).Msg("Locked contest")
	return true
}

func (w *WorkerImpl) unlockContest(id uint) bool {
	if err := w.dotlanLock.Unlock(w.ctx, strconv.Itoa(int(id))); err != nil {
		log.Warn().Uint("contest", id).Err(err).Msg("Error unlocking contest")
		return false
	}
	log.Trace().Uint("contest", id).Msg("Unlocked contest")
	return true
}

func (w *WorkerImpl) publishContest(messageType messagebroker.MessageTypes, subType messagebroker.MatchEvent, contest *database.DotlanStatus) error {

	if _, err := w.dbClient.Upsert(contest); err != nil {
		log.Warn().Err(err).Msg("Error happend while updating document")
		return err
	}

	msg := messagebroker.Message{
		Type:    messageType,
		SubType: subType.String(),
		Data:    &contest.MatchInfo,
	}

	if j, err := jsoniter.Marshal(msg); err != nil {
		log.Warn().Err(err).Msg("Error while marshalling message")
		return err
	} else {
		msg := message.Message{
			Payload: j,
		}

		err = w.messagePublisher.Publish(messagebroker.TOPIC, &msg)
		if err != nil {
			log.Error().Err(err).Msg("Error publishing to messagebroker")
			return err
		}
	}
	return nil
}

// processTournament processes all contests of a tournament
func (w *WorkerImpl) processTournament(tournament dotlan.Tournament, dotlanStates *database.DotlanStatusList) {
	log := log.With().Uint("tournament", tournament.Tid).Logger()

	if tournament.Tpause {
		// skip further processing since tournament is paused
		log.Debug().Msg("Skip tournament processing due to pause")
		return
	}

	// get all contests of the tournament
	contests, err := w.dotlanClient.GetContestForTournament(&tournament)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching contests for tournament")
		return
	}

	// process all contests
	for _, contest := range contests {
		log = log.With().Uint("contest", contest.Tcid).Logger()
		if !w.lockContest(contest.Tcid) {
			log.Info().Msg("Cannot lock contest, skipping")
			continue
		}
		defer w.unlockContest(contest.Tcid)

		if dotlanState := dotlanStates.GetElementById(contest.Tcid); dotlanState == nil {
			// new contest
			if tournament.Tclosed || tournament.Tpause {
				// skip contest for further processing since it's a new contest which is closed or paused
				log.Debug().Bool("Tclosed", tournament.Tclosed).Bool("Tpause", tournament.Tpause).Msg("Skipped contest")
				continue
			}
			log.Info().Msg("New contest found")

			// TODO: determine team name and playerinfos from dotlan into match info
			team1 := matchservice.Team{
				Id:    strconv.Itoa(contest.Team_a),
				Ready: contest.Ready_a.After(time.Time{}),
			}
			team2 := matchservice.Team{
				Id:    strconv.Itoa(contest.Team_b),
				Ready: contest.Ready_b.After(time.Time{}),
			}

			ds := database.DotlanStatus{
				DotlanContestID: contest.Tcid,
				Events: []database.DotlanStatusEvent{
					database.CMS_CONTEST_NEW,
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				MatchInfo: matchservice.MatchInfo{
					Id:           ksuid.New().String(),
					MsID:         strconv.Itoa(int(contest.Tcid)),
					Team1:        team1,
					Team2:        team2,
					PlayerAmount: 10,
					Game:         "csgo",
					Map:          "",
				},
			}

			go func() {
				err = w.publishContest(messagebroker.MessageTypeCreated, messagebroker.UNWINDIA_MATCH_NEW, &ds)
				if err != nil {
					log.Error().Err(err).Msg("Error publishing contest")
				}
			}()

			continue

		} else {
			// existing contest, checking for changes
			log.Debug().Msg("Processing existing contest")

			var subType messagebroker.MatchEvent = -1

			if tournament.Tclosed && !dotlanState.Events.Contains(database.CMS_CONTEST_FINISHED) {
				// tournament is closed, finish all contests to free ressources
				log.Debug().Msg("Finish contest due to closed tournament")
				dotlanState.Events = append(dotlanState.Events, database.CMS_CONTEST_FINISHED)
				subType = messagebroker.UNWINDIA_MATCH_FINISHED
			} else if contest.Won > 0 && !dotlanState.Events.Contains(database.CMS_CONTEST_FINISHED) {
				// contest is finished by dotlan
				log.Debug().Msg("Finish contest by dotlan")
				dotlanState.Events = append(dotlanState.Events, database.CMS_CONTEST_FINISHED)
				subType = messagebroker.UNWINDIA_MATCH_FINISHED
			} else if contest.Ready_a.After(time.Time{}) && !dotlanState.Events.Contains(database.CMS_CONTEST_READY_1) {
				log.Debug().Msg("Dotlan Team A ready")
				dotlanState.Events = append(dotlanState.Events, database.CMS_CONTEST_READY_1)
				dotlanState.MatchInfo.Team1.Ready = true
				subType = messagebroker.UNWINDIA_MATCH_READY_A
			} else if contest.Ready_b.After(time.Time{}) && !dotlanState.Events.Contains(database.CMS_CONTEST_READY_2) {
				log.Debug().Msg("Dotlan Team B ready")
				dotlanState.Events = append(dotlanState.Events, database.CMS_CONTEST_READY_2)
				dotlanState.MatchInfo.Team2.Ready = true
				subType = messagebroker.UNWINDIA_MATCH_READY_B
			} else if contest.Ready_a.After(time.Time{}) && contest.Ready_b.After(time.Time{}) && !dotlanState.Events.Contains(database.CMS_CONTEST_READY_ALL) {
				// TODO: well this step could maybe done directly with first iteration when all teams are ready but it's done this ugly way for now...
				log.Debug().Msg("Dotlan ALL TEAMS ARE READY")
				dotlanState.Events = append(dotlanState.Events, database.CMS_CONTEST_READY_ALL)
				subType = messagebroker.UNWINDIA_MATCH_READY_ALL
			}

			if subType != -1 {
				dotlanState.UpdatedAt = time.Now()
				err = w.publishContest(messagebroker.MessageTypeUpdated, subType, dotlanState)
				if err != nil {
					log.Error().Err(err).Msg("Error publishing contest")
				}
			}

			continue
		}
	}
}
