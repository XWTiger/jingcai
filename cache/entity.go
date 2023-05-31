package cache

type FootBallGames struct {
	TotalCount     interface{} `json:"totalCount"`
	LastUpdateTime interface{} `json:"lastUpdateTime"`
	CacheId        string
	MatchInfoList  []struct {
		BusinessDate string  `json:"businessDate"`
		MatchCount   int     `json:"matchCount"`
		Weekday      string  `json:"weekday"`
		SubMatchList []Match `json:"subMatchList"`
	} `json:"matchInfoList"`
	MatchDateList []struct {
		BusinessDate   string `json:"businessDate"`
		BusinessDateCn string `json:"businessDateCn"`
	} `json:"matchDateList"`
	LeagueList []struct {
		LeagueId       interface{} `json:"leagueId"`
		LeagueName     interface{} `json:"leagueName"`
		LeagueNameAbbr interface{} `json:"leagueNameAbbr"`
	} `json:"leagueList"`
}

type Match struct {
	MatchNum          int    `json:"matchNum"`
	HomeTeamCode      string `json:"homeTeamCode"`
	AwayTeamCode      string `json:"awayTeamCode"`
	MatchNumStr       string `json:"matchNumStr"`
	MatchWeek         string `json:"matchWeek"`
	LeagueId          int    `json:"leagueId"`
	LeagueCode        string `json:"leagueCode"`
	LeagueAbbName     string `json:"leagueAbbName"`
	LeagueAllName     string `json:"leagueAllName"`
	HomeTeamId        int    `json:"homeTeamId"`
	HomeTeamAbbName   string `json:"homeTeamAbbName"`
	HomeTeamAllName   string `json:"homeTeamAllName"`
	HomeTeamAbbEnName string `json:"homeTeamAbbEnName"`
	AwayTeamId        int    `json:"awayTeamId"`
	AwayTeamAbbName   string `json:"awayTeamAbbName"`
	AwayTeamAllName   string `json:"awayTeamAllName"`
	AwayTeamAbbEnName string `json:"awayTeamAbbEnName"`
	MatchDate         string `json:"matchDate"`
	BusinessDate      string `json:"businessDate"`
	MatchTime         string `json:"matchTime"`
	MatchId           string `json:"matchId"`
	SellStatus        int    `json:"sellStatus"`
	MatchStatus       string `json:"matchStatus"`
	Remark            string `json:"remark"`
	IsHot             int    `json:"isHot"`
	IsHide            int    `json:"isHide"`
	HomeRank          string `json:"homeRank"`
	AwayRank          string `json:"awayRank"`
	OddsList          []struct {
		A          string      `json:"a"`
		Af         interface{} `json:"af"`
		Df         interface{} `json:"df"`
		Hf         interface{} `json:"hf"`
		D          string      `json:"d"`
		H          string      `json:"h"`
		PoolCode   string      `json:"poolCode"`
		PoolId     string      `json:"poolId"`
		UpdateDate interface{} `json:"updateDate"`
		UpdateTime interface{} `json:"updateTime"`
		GoalLine   string      `json:"goalLine"`
		Hrate      interface{} `json:"hrate"`
		Drate      interface{} `json:"drate"`
		Arate      interface{} `json:"arate"`
	} `json:"oddsList"`
	PoolList []Pool `json:"poolList"`
	Had      struct {
		A          string      `json:"a"`
		Af         string      `json:"af"`
		Df         string      `json:"df"`
		Hf         string      `json:"hf"`
		D          string      `json:"d"`
		H          string      `json:"h"`
		PoolCode   interface{} `json:"poolCode"`
		PoolId     interface{} `json:"poolId"`
		UpdateDate interface{} `json:"updateDate"`
		UpdateTime interface{} `json:"updateTime"`
		GoalLine   string      `json:"goalLine"`
		Hrate      string      `json:"hrate"`
		Drate      string      `json:"drate"`
		Arate      string      `json:"arate"`
	} `json:"had"`
	Hhad struct {
		A          string      `json:"a"`
		Af         string      `json:"af"`
		Df         string      `json:"df"`
		Hf         string      `json:"hf"`
		D          string      `json:"d"`
		H          string      `json:"h"`
		PoolCode   interface{} `json:"poolCode"`
		PoolId     interface{} `json:"poolId"`
		UpdateDate interface{} `json:"updateDate"`
		UpdateTime interface{} `json:"updateTime"`
		GoalLine   string      `json:"goalLine"`
		Hrate      string      `json:"hrate"`
		Drate      string      `json:"drate"`
		Arate      string      `json:"arate"`
	} `json:"hhad"`
	Crs struct {
		GoalLine   string `json:"goalLine"`
		S00S00     string `json:"s00s00"`
		S00S00F    string `json:"s00s00f"`
		S00S01     string `json:"s00s01"`
		S00S01F    string `json:"s00s01f"`
		S00S02     string `json:"s00s02"`
		S00S02F    string `json:"s00s02f"`
		S00S03     string `json:"s00s03"`
		S00S03F    string `json:"s00s03f"`
		S00S04     string `json:"s00s04"`
		S00S04F    string `json:"s00s04f"`
		S00S05     string `json:"s00s05"`
		S00S05F    string `json:"s00s05f"`
		S01S00     string `json:"s01s00"`
		S01S00F    string `json:"s01s00f"`
		S01S01     string `json:"s01s01"`
		S01S01F    string `json:"s01s01f"`
		S01S02     string `json:"s01s02"`
		S01S02F    string `json:"s01s02f"`
		S01S03     string `json:"s01s03"`
		S01S03F    string `json:"s01s03f"`
		S01S04     string `json:"s01s04"`
		S01S04F    string `json:"s01s04f"`
		S01S05     string `json:"s01s05"`
		S01S05F    string `json:"s01s05f"`
		S1Sa       string `json:"s1sa"`
		S1Saf      string `json:"s1saf"`
		S1Sd       string `json:"s1sd"`
		S1Sdf      string `json:"s1sdf"`
		S1Sh       string `json:"s1sh"`
		S1Shf      string `json:"s1shf"`
		S02S00     string `json:"s02s00"`
		S02S00F    string `json:"s02s00f"`
		S02S01     string `json:"s02s01"`
		S02S01F    string `json:"s02s01f"`
		S02S02     string `json:"s02s02"`
		S02S02F    string `json:"s02s02f"`
		S02S03     string `json:"s02s03"`
		S02S03F    string `json:"s02s03f"`
		S02S04     string `json:"s02s04"`
		S02S04F    string `json:"s02s04f"`
		S02S05     string `json:"s02s05"`
		S02S05F    string `json:"s02s05f"`
		S03S00     string `json:"s03s00"`
		S03S00F    string `json:"s03s00f"`
		S03S01     string `json:"s03s01"`
		S03S01F    string `json:"s03s01f"`
		S03S02     string `json:"s03s02"`
		S03S02F    string `json:"s03s02f"`
		S03S03     string `json:"s03s03"`
		S03S03F    string `json:"s03s03f"`
		S04S00     string `json:"s04s00"`
		S04S00F    string `json:"s04s00f"`
		S04S01     string `json:"s04s01"`
		S04S01F    string `json:"s04s01f"`
		S04S02     string `json:"s04s02"`
		S04S02F    string `json:"s04s02f"`
		S05S00     string `json:"s05s00"`
		S05S00F    string `json:"s05s00f"`
		S05S01     string `json:"s05s01"`
		S05S01F    string `json:"s05s01f"`
		S05S02     string `json:"s05s02"`
		S05S02F    string `json:"s05s02f"`
		UpdateDate string `json:"updateDate"`
		UpdateTime string `json:"updateTime"`
	} `json:"crs"`
	Ttg struct {
		GoalLine   string `json:"goalLine"`
		S0         string `json:"s0"`
		S0F        string `json:"s0f"`
		S1         string `json:"s1"`
		S1F        string `json:"s1f"`
		S2         string `json:"s2"`
		S2F        string `json:"s2f"`
		S3         string `json:"s3"`
		S3F        string `json:"s3f"`
		S4         string `json:"s4"`
		S4F        string `json:"s4f"`
		S5         string `json:"s5"`
		S5F        string `json:"s5f"`
		S6         string `json:"s6"`
		S6F        string `json:"s6f"`
		S7         string `json:"s7"`
		S7F        string `json:"s7f"`
		UpdateDate string `json:"updateDate"`
		UpdateTime string `json:"updateTime"`
	} `json:"ttg"`
	Hafu struct {
		Aa         string `json:"aa"`
		Aaf        string `json:"aaf"`
		Ad         string `json:"ad"`
		Adf        string `json:"adf"`
		Ah         string `json:"ah"`
		Ahf        string `json:"ahf"`
		Da         string `json:"da"`
		Daf        string `json:"daf"`
		Dd         string `json:"dd"`
		Ddf        string `json:"ddf"`
		Dh         string `json:"dh"`
		Dhf        string `json:"dhf"`
		GoalLine   string `json:"goalLine"`
		Ha         string `json:"ha"`
		Haf        string `json:"haf"`
		Hd         string `json:"hd"`
		Hdf        string `json:"hdf"`
		Hh         string `json:"hh"`
		Hhf        string `json:"hhf"`
		Id         int    `json:"id"`
		UpdateDate string `json:"updateDate"`
		UpdateTime string `json:"updateTime"`
	} `json:"hafu"`
}

type Pool struct {
	MatchId           int    `json:"matchId"`
	MatchNum          int    `json:"matchNum"`
	PoolId            int    `json:"poolId"`
	PoolCode          string `json:"poolCode"`
	PoolOddsType      string `json:"poolOddsType"`
	PoolStatus        string `json:"poolStatus"`
	FixedOddsgoalLine string `json:"fixedOddsgoalLine"`
	PoolCloseDate     string `json:"poolCloseDate"`
	PoolCloseTime     string `json:"poolCloseTime"`
	SellInitialDate   string `json:"sellInitialDate"`
	SellInitialTime   string `json:"sellInitialTime"`
	CbtAllUp          int    `json:"cbtAllUp"`
	CbtSingle         int    `json:"cbtSingle"`
	CbtValue          int    `json:"cbtValue"`
	IntAllUp          int    `json:"intAllUp"`
	IntSingle         int    `json:"intSingle"`
	IntValue          int    `json:"intValue"`
	VbtAllUp          int    `json:"vbtAllUp"`
	VbtSingle         int    `json:"vbtSingle"`
	VbtValue          int    `json:"vbtValue"`
	BettingSingle     int    `json:"bettingSingle"`
	BettingAllup      int    `json:"bettingAllup"`
	Single            int    `json:"single"`
	AllUp             int    `json:"allUp"`
	UpdateDate        string `json:"updateDate"`
	UpdateTime        string `json:"updateTime"`
}

func (f *FootBallGames) MatchListToMap() map[string]Match {
	var mapper = make(map[string]Match)
	for _, s := range f.MatchInfoList {
		for _, match := range s.SubMatchList {
			mapper[match.MatchId] = match
		}
	}
	return mapper
}

func (f *FootBallGames) GetPoolMap() map[int]Pool {
	var mapper = make(map[int]Pool)
	for _, s := range f.MatchInfoList {
		for _, match := range s.SubMatchList {
			if len(match.PoolList) > 0 {
				for _, pool := range match.PoolList {
					_, ok := mapper[pool.PoolId]
					if !ok {
						mapper[pool.PoolId] = pool
					}
				}
			}
		}
	}
	return mapper
}

func (f *FootBallGames) GetSinglePoolMap() map[int]Pool {
	var mapper = make(map[int]Pool)
	for _, s := range f.MatchInfoList {
		for _, match := range s.SubMatchList {
			if len(match.PoolList) > 0 {
				for _, pool := range match.PoolList {
					_, ok := mapper[pool.PoolId]
					if !ok {
						mapper[pool.PoolId] = pool
					}
				}
			}
		}
	}
	return mapper
}
