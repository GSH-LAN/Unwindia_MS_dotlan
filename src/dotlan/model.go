package dotlan

import (
	"database/sql"
	"time"
)

type TournamentsResult struct {
	Result TournamentList
	Error  error
}

type TournamentList []Tournament

type Tournament struct {
	Tid      uint   `db:"tid" `
	Tgroupid int    `db:"tgroupid" `
	Tactive  int    `db:"tactive" `
	Topen    bool   `db:"topen" `
	Tclosed  bool   `db:"tclosed" `
	Tpause   bool   `db:"tpause" `
	Teventid int    `db:"teventid" `
	Tname    string `db:"tname" `
	// Tplaytype    string         `db:"tplaytype" `
	// Tparams      string         `db:"tparams" `
	// Tminanz      sql.NullInt64  `db:"tminanz" `
	// Tmaxanz      sql.NullInt64  `db:"tmaxanz" `
	// Tuserproteam int            `db:"tuserproteam" `
	// Tmoreplayer  bool           `db:"tmoreplayer" `
	// Tlogo        sql.NullString `db:"tlogo" `
	// Tregeln      sql.NullString `db:"tregeln" `
	// Tinfotext    sql.NullString `db:"tinfotext" `
	// Tstart string `db:"tstart" `
	// Troundtime            int            `db:"troundtime" `
	// Troundpause           int            `db:"troundpause" `
	// Tcheckin              bool           `db:"tcheckin" `
	// Tcheckintime          int            `db:"tcheckintime" `
	// Tautodefaultwin       bool           `db:"tautodefaultwin" `
	// Taudultsonly          bool           `db:"tadultsonly" `
	// Tpassword             string         `db:"tpassword" `
	// Tprice1               string         `db:"tprice1" `
	// Tprice2               string         `db:"tprice2" `
	// Tprice3               string         `db:"tprice3" `
	// Tnight                bool           `db:"tnight" `
	// Tnightbegin           string         `db:"tnightbegin" `
	// Tnightend             string         `db:"tnightend" `
	Tdefmap sql.NullString `db:"tdefmap" `
	// Tadmins               sql.NullString `db:"tadmins" `
	// Tcoins                int            `db:"tcoins" `
	// Tcoinsreturn          bool           `db:"tcoinsreturn" `
	// C_contests            int            `db:"c_contests" `
	// C_contests_complete   int            `db:"c_contests_complete" `
	// C_teilnehmer          int            `db:"c_teilnehmer" `
	// C_teilnehmer_complete int            `db:"c_teilnehmer_complete" `
	// C_complete            int            `db:"c_complete" `
	// C_round               string         `db:"c_round" `
	// Tname_short           string         `db:"tname_short" ` //tname_short custom flag by GSH: ?
	// Tgameserver           bool           `db:"tgameserver" ` //tgameserver custom flag by GSH: gameserver management by UNWINDIA enabled
}

func (Tournament) TableName() string {
	return "t_turnier"
}

type ContestsResult struct {
	Result ContestList
	Error  error
}

type ContestList []Contest

type Contest struct {
	Tcid uint `db:"tcid"`
	Tid  uint `db:"tid"`
	//Tcrunde    int            `db:"tcrunde"`
	Team_a int `db:"team_a"`
	Team_b int `db:"team_b"`
	//Wins_a     int            `db:"wins_a"`
	//Wins_b     int            `db:"wins_b"`
	Won int `db:"won"`
	//Dateline   time.Time      `db:"dateline"`
	Host string `db:"host"`
	//User_id    int            `db:"user_id"`
	//Row        int            `db:"row"`
	//Comments   int            `db:"comments"`
	//Starttime  time.Time      `db:"starttime"`
	//Ignoretime int            `db:"ignoretime"`
	Ready_a time.Time `db:"ready_a"`
	Ready_b time.Time `db:"ready_b"`
	//Defaultwin int            `db:"defaultwin"`
	//Intern     sql.NullString `db:"intern"`
	//Public     sql.NullString `db:"public"`
	Tournament Tournament
}

func (Contest) TableName() string {
	return "t_contest"
}

type ParticipantsTeam struct {
	tnid          int
	tid           int
	tnname        sql.NullString
	tnanz         sql.NullInt64
	tnleader      sql.NullInt64
	tnpasswd      sql.NullString
	tcheckedin    int
	loser         int
	tnlogo        int
	homepage      string
	irc           string
	description   sql.NullString
	countrycode   string
	rankpos       string
	handy         string
	orgakommentar string
}
