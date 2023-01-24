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

type Team struct {
	Tnid     int            `db:"tnid"`
	Tid      int            `db:"tid"`
	Tnname   sql.NullString `db:"tnname"`
	Tnanz    sql.NullInt64  `db:"tnanz"`
	Tnleader sql.NullInt64  `db:"tnleader"`
	//Tnpasswd sql.NullString `db:"tnpasswd"`
	//Tcheckedin    int            `db:"tcheckedin"`
	//Loser         int            `db:"loser"`
	Tnlogo int `db:"tnlogo"`
	//Homepage      string         `db:"homepage"`
	//Irc           string         `db:"irc"`
	//Description   sql.NullString `db:"description"`
	//Countrycode   string         `db:"countrycode"`
	//Rankpos       string         `db:"rankpos"`
	//Handy         string         `db:"handy"`
	//Orgakommentar string         `db:"orgakommentar"`
}

func (Team) TableName() string {
	return "t_teilnehmer"
}

type ParticipantsTeamList []ParticipantsTeam

type ParticipantsTeam struct {
	Tnid   int `db:"tnid"`
	UserId int `db:"user_id"`
	//Dateline time.Time `db:"dateline"`
}

func (ParticipantsTeam) TableName() string {
	return "t_teilnehmer_part"
}

type UserList []User

type User struct {
	ID   int    `db:"id"`
	Nick string `db:"nick"`
	//passwort            string    `db:"passwort"`
	//clan                string    `db:"clan"`
	//clantag             string    `db:"clantag"`
	//vorname             string    `db:"vorname"`
	//nachname            string    `db:"nachname"`
	//strasse             string    `db:"strasse"`
	//plz                 string    `db:"plz"`
	//wohnort             string    `db:"wohnort"`
	//countrycode         string    `db:"countrycode"`
	//telefon             string    `db:"telefon"`
	//handy               string    `db:"handy"`
	//fax                 string    `db:"fax"`
	//geb                 time.Time `db:"geb"`
	//email               string    `db:"email"`
	//newsletter          bool      `db:"newsletter"`
	//female              bool      `db:"female"`
	//skype               string    `db:"skype"`
	//homepage            string    `db:"homepage"`
	//motto               string    `db:"motto"`
	//ver_age             int       `db:"ver_age"`
	//kommentar           string    `db:"kommentar"`
	//joinus              int       `db:"joinus"`
	//ver_email           int       `db:"ver_email"`
	//ver_salt            string    `db:"ver_salt"`
	//lastpost            int       `db:"lastpost"`
	//lastgetpasswd       int       `db:"lastgetpasswd"`
	//style               string    `db:"style"`
	//lastactivity        time.Time `db:"lastactivity"`
	//lastupdate          int       `db:"lastupdate"`
	//last_ip             string    `db:"last_ip"`
	//lastvisit_forum     time.Time `db:"lastvisit_forum"`
	//lastvisit_groupware int       `db:"lastvisit_groupware"`
	//settings            string    `db:"settings"`
	Avatar bool `db:"avatar"`
	//avatar_name string `db:"avatar_name"`
	//external_id         string    `db:"external_id"`
	//external_info       string    `db:"external_info"`
	//stats_postings      int       `db:"stats_postings"`
	//stats_comments      int       `db:"stats_comments"`
	//stats_messages_recv int       `db:"stats_messages_recv"`
	//stats_messages_sent int       `db:"stats_messages_sent"`
	//stats_sessions      int       `db:"stats_sessions"`
	//stats_hits          int       `db:"stats_hits"`
	//stats_profile_hits  int       `db:"stats_profile_hits"`
	//kennzeichen         string    `db:"kennzeichen"`
	//discord_id          string    `db:"discord_id"`
	//pubg_name           string    `db:"pubg_name"`
	//epic_name           string    `db:"epic_name"`
}

func (User) TableName() string {
	return "user"
}
