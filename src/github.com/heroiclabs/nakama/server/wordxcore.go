package server

import (
	"database/sql"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/jsonpb"
	"github.com/tbalthazar/onesignal-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"sync"
	"time"
)

type HookCallback = func(logger *zap.Logger, config Config, runtimePool *RuntimePool, jsonpbMarshaler *jsonpb.Marshaler, jsonpbUnmarshaler *jsonpb.Unmarshaler, sessionID string, uid uuid.UUID, username string, expiry int64, req interface{}) (interface{}, error)
type HookCallbackRPC = func(queryParams map[string][]string, uid string, username string, sessionExpiry int64, sid string, payload interface{}) (interface{}, error, codes.Code)
type HookCallbackBefore = HookCallback
type HookCallbackAfter = func(logger *zap.Logger, config Config, runtimePool *RuntimePool, jsonpbMarshaler *jsonpb.Marshaler, sessionID string, uid uuid.UUID, username string, expiry int64, req interface{})
type HookCallbackMatchmaking = func(logger *zap.Logger, runtimePool *RuntimePool, entries []*MatchmakerEntry) (err error, matchId *string)

type WXCallbacks struct {
	RPC        map[string]HookCallbackRPC
	Before     map[string]HookCallbackBefore
	After      map[string]HookCallbackAfter
	Matchmaker HookCallbackMatchmaking
}

type DBCleaner struct {
	db       *sql.DB
	interval time.Duration
	prevTick time.Time
	wx       *WordX
}

type TCron struct {
	date  time.Time
	round *Tour
}

type TournamentCron struct {
	db       *sql.DB
	interval time.Duration
	prevTick time.Time
	wx       *WordX
	mu       *sync.Mutex
	crons    map[uuid.UUID]*TCron
}

type WordX struct {
	Callbacks        WXCallbacks
	Logger           *zap.Logger
	db               *sql.DB
	Cache            *WordXCache
	apiServ          *ApiServer
	cachedCategories *CategoriesList
	MM               *WordXMM
	TournamentCron   *TournamentCron
	Dbcleaner        *DBCleaner
	oneSignal        *onesignal.Client
	oneSignalApps    []onesignal.App
	oneSignalAppID   string
}

func (d *TournamentCron) AddCron(t *Tournament) {
	var date time.Time
	var round *Tour
	if t.CurrentRound == nil {
		date = *t.StartDate
	} else {
		date = t.CurrentRound.EndDate
		copyR := *t.CurrentRound
		round = &copyR
	}

	d.mu.Lock()
	d.crons[t.Id] = &TCron{
		date:  date,
		round: round,
	}
	d.mu.Unlock()
}

func (d *TournamentCron) RunRoutine(logger *zap.Logger) {
	logger.Info("Tournament cron is running")

	ticker := d.updateTicker()
	for {
		<-ticker.C
		ticker = d.updateTicker()
	}
}

func (d *TournamentCron) cronTournament(tid uuid.UUID, t *Tournament) {
	if t.CurrentRound == nil {
		d.wx.Logger.Info("Starting tournament <" + tid.String() + ">")
		//make tournament active
		//send notifications

		round, err := t.GetNextRound()
		if err != nil {
			d.wx.Logger.Error("cronTournament: ", zap.Error(err))
			return
		}
		d.wx.apiServ.setRound(round, t)

	} else {
		d.wx.Logger.Info("Round <" + tid.String() + "> ended")

		round, err := t.GetNextRound()
		if err != nil {
			d.wx.Logger.Error("cronTournament: ", zap.Error(err))
			return
		}

		d.wx.apiServ.setRound(round, t)
	}
}

func (d *TournamentCron) updateTicker() *time.Ticker {
	_, tt, err := d.wx.apiServ.GetTournamentsInner(nil, true, -1)
	for _, t := range tt {
		tournament := t
		d.wx.Cache.Tournaments[t.Id] = tournament
	}

	if err != nil {
		d.wx.Logger.Error("TournamentCron. update ticker error", zap.Error(err))
	}

	if tt != nil {
		err = d.wx.apiServ.GetTournamentsRounds(tt)
		if err != nil {
			d.wx.Logger.Error("TournamentCron. update ticker error", zap.Error(err))
		}
	}

	for _, t := range tt {
		var date time.Time
		if t.CurrentRound != nil {
			date = t.CurrentRound.EndDate
		} else {
			date = *t.StartDate
		}

		if (t.Status == TS_PUB || t.Status == TS_ACT) && time.Now().After(date) {
			d.cronTournament(t.Id, t)
		}
	}

	return time.NewTicker(d.interval)
}

func (d *DBCleaner) RunRoutine(logger *zap.Logger) {
	logger.Info("DB Cleaning routine is running")

	ticker := d.updateTicker()
	for {
		<-ticker.C
		ticker = d.updateTicker()
	}
}

func (d *DBCleaner) doCleaning() {
	d.db.Exec("delete from notification WHERE create_time < NOW() - '7 days'::interval;")
	d.db.Exec("delete from matches WHERE last_turn < NOW() - '7 days'::interval;")

	d.wx.MM.ResolveOldMatches()
}

func (d *DBCleaner) updateTicker() *time.Ticker {

	if time.Now().After(d.prevTick.Add(d.interval)) {
		d.prevTick = time.Now()
		d.doCleaning()
	}

	return time.NewTicker(d.interval)
}

type WordXCache struct {
	Tournaments map[uuid.UUID]*Tournament
	Settings    *Settings
	CatMap      map[uuid.UUID]*Category
	MMWordsPool *Category
	wx          *WordX
}

func (c *WordXCache) GetImage(id string) (error, *string) {
	var uid, err = uuid.FromString(id)
	if err != nil {
		return err, nil
	} else {
		if cat, ok := c.CatMap[uid]; ok {
			return nil, &cat.ImageData
		} else {
			if t, ok := c.Tournaments[uid]; ok {
				return nil, &t.ImageData
			} else {
				return status.Error(codes.Internal, "No such category Id"), nil
			}
		}
	}
}

func (wx *WordX) HasCallback(mode ExecutionMode, id string) (ok bool) {
	switch mode {
	case ExecutionModeRPC:
		_, ok = wx.Callbacks.RPC[id]
	case ExecutionModeBefore:
		_, ok = wx.Callbacks.Before[id]
	case ExecutionModeAfter:
		_, ok = wx.Callbacks.After[id]
	case ExecutionModeMatchmaker:
		ok = wx.Callbacks.Matchmaker != nil
	}

	return ok
}

func CreateWordX(logger *zap.Logger, db *sql.DB) *WordX {
	wx := &WordX{
		Callbacks: WXCallbacks{
			Before: make(map[string]HookCallbackBefore),
			After:  make(map[string]HookCallbackAfter),
			RPC:    make(map[string]HookCallbackRPC),
		},
		Logger:    logger,
		db:        db,
		oneSignal: onesignal.NewClient(nil),
	}
	wx.MM = CreateWordXMM(db)
	wx.MM.wx = wx
	wx.oneSignal.UserKey = "Mzg0OTc2ZmEtZmZmMi00ZjBmLTlmYTctNDIwNjIzNDdlZTIw"
	wx.oneSignal.AppKey = "MjcwMTI0MTctODc0MS00YzQ0LWJlNzItODVkNTAyMWZlOWY1"

	apps, _, err := wx.oneSignal.Apps.List()
	if err != nil {
		logger.DPanic("Can't get oneSignal apps")
		panic(err)
	}
	//playerid := "76fb5a5f-9c31-4820-9e4c-24f0838428f2"

	for _, a := range apps {
		wx.oneSignalAppID = a.ID
	}

	if wx.oneSignalAppID == "" {
		logger.DPanic("Can't find oneSignal wordx app")
	}

	//	uid := uuid.Must(uuid.FromString("3a76b16a-fae6-4378-a07c-1c43d7b9bf1e"))

	//	wx.SendMassPushTo([]uuid.UUID{uid}, "test push", "content")

	InitWordXAPI(wx)

	return wx
}

func (wx *WordX) Matchmaking(logger *zap.Logger, runtimePool *RuntimePool, entries []*MatchmakerEntry) (err error, matchid *string) {
	println(entries)
	return nil, nil
}

func (wx *WordX) CacheCategories(c *WordXCache) {
	query := "SELECT id, catname, words_eng, words_spa, words_ita, words_fre, words_por, points_eng, points_spa, points_ita, points_fre, points_por, description, unlock_price, image_data from category;"
	rows, err := wx.apiServ.db.Query(query)
	if err != nil {
		panic(err)
	}

	c.CatMap = make(map[uuid.UUID]*Category)

	for rows.Next() {
		cat, err := ScanCat(rows, nil, true)
		if err != nil {
			panic(err)
		}
		c.CatMap[cat.Id] = cat
	}

	c.MMWordsPool = &Category{}

	DefLangs(c.MMWordsPool)
	for _, x := range c.CatMap {
		for inx, l := range x.Languages {
			for wordinx, _ := range l.Words {
				c.MMWordsPool.Languages[inx].Words = append(c.MMWordsPool.Languages[inx].Words, l.Words[wordinx])
				c.MMWordsPool.Languages[inx].Points = append(c.MMWordsPool.Languages[inx].Points, l.Points[wordinx])
			}
		}
	}

	println()
}

func (wx *WordX) Init(api *ApiServer) {
	wx.apiServ = api
	wx.Cache = &WordXCache{
		wx: wx,
	}

	wx.Cache.Settings, _ = wx.apiServ.GetSettingsInner()
	wx.Cache.Tournaments = make(map[uuid.UUID]*Tournament)

	wx.CacheCategories(wx.Cache)
	wx.Callbacks.Matchmaker = wx.Matchmaking
	wx.Dbcleaner = &DBCleaner{
		db:       wx.db,
		interval: time.Hour,
		wx:       wx,
	}

	wx.TournamentCron = &TournamentCron{
		mu:       &sync.Mutex{},
		interval: time.Minute,
		crons:    make(map[uuid.UUID]*TCron),
		wx:       wx,
	}

	_, tt, _ := wx.apiServ.GetTournamentsInner(nil, true, -1)
	for inx, x := range tt {
		t := tt[inx]
		wx.Cache.Tournaments[x.Id] = t
	}
}

func (wx *WordX) AddCallbackBefore(cbname string, cb HookCallbackBefore) {
	wx.Callbacks.Before[strings.ToLower(cbname)] = cb
}

func (wx *WordX) AddCallbackAfter(cbname string, cb HookCallbackAfter) {
	wx.Callbacks.After[strings.ToLower(cbname)] = cb
}

func (wx *WordX) AddCallbackRPC(cbname string, cb HookCallbackRPC) {
	wx.Callbacks.RPC[strings.ToLower(cbname)] = cb
}

func (wx *WordX) BeforeHook(logger *zap.Logger, config Config, runtimePool *RuntimePool, jsonpbMarshaler *jsonpb.Marshaler, jsonpbUnmarshaler *jsonpb.Unmarshaler, sessionID string, uid uuid.UUID, username string, expiry int64, callbackID string, req interface{}) (interface{}, error) {
	if fnc, ok := wx.Callbacks.Before[strings.ToLower(callbackID)]; ok {
		fnc(logger, config, runtimePool, jsonpbMarshaler, jsonpbUnmarshaler, sessionID, uid, username, expiry, req)
	}

	return req, nil
}

func (wx *WordX) AfterHook(logger *zap.Logger, config Config, runtimePool *RuntimePool, jsonpbMarshaler *jsonpb.Marshaler, sessionID string, uid uuid.UUID, username string, expiry int64, callbackID string, req interface{}) {
	if fnc, ok := wx.Callbacks.After[strings.ToLower(callbackID)]; ok {
		fnc(logger, config, runtimePool, jsonpbMarshaler, sessionID, uid, username, expiry, req)
	}
}

func (wx *WordX) InvokeRPCFunction(cbname string, queryParams map[string][]string, uid string, username string, sessionExpiry int64, sid string, payload interface{}) (result interface{}, err error, code codes.Code) {
	if fnc, ok := wx.Callbacks.RPC[strings.ToLower(cbname)]; ok {
		result, err, code = fnc(queryParams, uid, username, sessionExpiry, sid, payload)
	} else {
		err = status.Error(codes.Internal, "RPC function not found")
		code = codes.InvalidArgument
	}

	return
}
