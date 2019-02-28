package server

import (
	"database/sql"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/labstack/gommon/random"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math"
	"math/rand"
	"sync"
	"time"
)

const (
	INV_STATUS_SENT      = "sent"
	INV_STATUS_DELIVERED = "delivered"
)

const (
	CT_BOTH     = "both"
	CT_YOU      = "you"
	CT_OPPONENT = "opponent"
)

type MatchStatus = string
type UUIDPair = string

const (
	MS_INVITING = "inviting"
	MS_INVITED  = "invited"
	MS_PROGRESS = "in_progress"
	MS_FINISHED = "finished"
)

const (
	PS_WIN  = "win"
	PS_LOSE = "lose"
	PS_DRAW = "draw"
)

type Prize struct {
	Status      string
	AddedPoints int
	AddedCoins  int
}

type RoundWords struct {
	Round1 Words
	Round2 Words
	Round3 Words
}

type MatchToClient struct {
	Id            uuid.UUID
	OpponentID    uuid.UUID
	WaitingUser   string
	Round         int
	Status        MatchStatus
	InviteId      uuid.UUID `json:"-"`
	LastTurnTime  time.Time
	YourScore     []int64
	OpponentScore []int64
	InnerStatus   MatchStatus `json:"-"`
	User1Score    []int64     `json:"-"`
	User2Score    []int64     `json:"-"`
	WinUser       *uuid.UUID  `json:"-"`
	User1         uuid.UUID   `json:"-"`
	User2         uuid.UUID   `json:"-"`
	WaitingUserID *uuid.UUID  `json:"-"`
	Opponent      *Opponent
	Prize         *Prize
	Words         *RoundWords
	Seed          int64 `json:"-"`
}

func (w *WordXMM) ResolveOldMatches() error {
	rows, err := w.db.Query(`SELECT id, user1, user2, waitingUser, round, status, invite_id, last_turn, user1_score, user2_score, win_user, words_seed from matches WHERE status != 'finished' AND last_turn < NOW() - '3 days'::interval`)
	if err != nil {
		return err
	}

	if rows.Next() {
		err, match := w.ScanMatch(rows, uuid.Must(uuid.NewV4()), false)
		if err == nil {
			if match.WaitingUserID == nil {
				match.InnerStatus = MS_FINISHED
				match.WinUser = nil
				w.SaveMatch(match)
				var zeroUUID uuid.UUID
				w.wx.apiServ.NotifyYourTurn(match, zeroUUID, 10)
			} else {
				waiting := *match.WaitingUserID
				var winUser uuid.UUID
				if match.User1 == waiting {
					winUser = match.User2
				} else {
					winUser = match.User1
				}

				match.WinUser = &winUser
				match.InnerStatus = MS_FINISHED
				w.SaveMatch(match)
				var zeroUUID uuid.UUID
				w.wx.apiServ.NotifyYourTurn(match, zeroUUID, 10)
			}
		}

	} else {
		return status.Error(codes.InvalidArgument, "Can't scan match")
	}

	if err != nil {
		return err
	}

	return nil
}

func (w *WordXMM) GetWordsForRound(round int, seed int64) Words {
	var inserted = make(map[int]bool)
	var wlist Words
	const TRY_NUM = 2000
	for tries := 0; tries < TRY_NUM; tries++ {
		rand.Seed(seed)
		seed++
		langInx := rand.Intn(len(w.wx.Cache.MMWordsPool.Languages))

		ww := w.wx.Cache.MMWordsPool.Languages[langInx].Words
		if ww == nil {
			return wlist
		}

		intn := rand.Intn(len(ww))

		if b, ok := inserted[intn]; !ok || !b {
			l := w.wx.Cache.MMWordsPool.Languages[langInx]
			if l.Words[intn] == "" {
				continue
			}

			wlist.Words = append(wlist.Words, l.Words[intn])
			wlist.Points = append(wlist.Points, l.Points[intn])

			inserted[intn] = true

			if len(wlist.Words) >= int(w.wx.Cache.Settings.H2HWords) {
				break
			}
		}

	}

	wlist.Awards = AwardsFromWords(wlist.Words)

	return wlist
}

func (w *WordXMM) UpdateWords(m *MatchToClient, seed int64) {

	m.Words = &RoundWords{}
	m.Words.Round1 = w.GetWordsForRound(1, seed)
	m.Words.Round2 = w.GetWordsForRound(2, seed+1000)
	m.Words.Round3 = w.GetWordsForRound(3, seed+2000)
}

func (match *MatchToClient) toString(me uuid.UUID, wx *WordX) string {
	var m = *match
	m.Update(me, wx)
	m.Prize = nil
	m.Words = nil
	bytea, _ := json.Marshal(m)
	return string(bytea)
}

func (match *MatchToClient) Update(me uuid.UUID, wx *WordX) {

	if match.WaitingUserID == nil {
		match.WaitingUser = CT_BOTH
	} else {
		if *match.WaitingUserID == me {
			match.WaitingUser = CT_YOU
		} else {
			match.WaitingUser = CT_OPPONENT
		}
	}
	if match.InnerStatus == MS_INVITED {
		if match.User1 == me {
			match.Status = MS_INVITING
		} else {
			match.Status = MS_INVITED
		}
	} else {
		match.Status = match.InnerStatus
	}

	var yourScore, opponentScore []int64
	var opponent uuid.UUID

	if me == match.User1 {
		yourScore = match.User1Score
		opponent = match.User2
		opponentScore = match.User2Score
	} else {
		yourScore = match.User2Score

		opponent = match.User1
		opponentScore = match.User1Score
	}

	match.OpponentID = opponent
	match.YourScore = yourScore
	match.OpponentScore = opponentScore

	if match.InnerStatus == MS_FINISHED {
		if match.WinUser == nil {
			match.Prize = &Prize{
				AddedPoints: 0,
				AddedCoins:  0,
				Status:      PS_DRAW,
			}
		} else {
			if *match.WinUser == me {
				var summ int = 0
				for _, x := range match.YourScore {
					summ += int(x)
				}

				match.Prize = &Prize{
					AddedPoints: summ,
					AddedCoins:  int(float32(summ) * wx.Cache.Settings.CoinsPerPoint),
					Status:      PS_WIN,
				}

			} else {
				match.Prize = &Prize{
					AddedPoints: 0,
					AddedCoins:  0,
					Status:      PS_LOSE,
				}
			}
		}
	}
}

func (match *MatchToClient) NextTurn(me uuid.UUID, points int, wx *WordX) {
	if match.User1 == me {
		match.User1Score = append(match.User1Score, int64(points))
	} else {
		match.User2Score = append(match.User2Score, int64(points))
	}
	var gotWin bool

	if match.WaitingUserID == nil {
		match.WaitingUserID = &match.OpponentID
	} else {
		if *match.WaitingUserID == me {
			match.Round++
			match.WaitingUserID = nil
			if match.Round == 4 {
				match.Round = 3
				summUser1 := 0
				summUser2 := 0
				for _, x := range match.User1Score {
					summUser1 += int(x)
				}

				for _, x := range match.User2Score {
					summUser2 += int(x)
				}

				if summUser1 == summUser2 {
					match.WinUser = nil
				} else {
					if summUser1 > summUser2 {
						match.WinUser = &match.User1
					} else {
						match.WinUser = &match.User2
					}
				}

				gotWin = true
				match.InnerStatus = MS_FINISHED
			}
		}
	}

	match.LastTurnTime = time.Now()
	match.Update(me, wx)

	if gotWin {
		if match.WinUser != nil && match.Prize != nil {
			id := *match.WinUser
			wx.apiServ.AddResource(id, int32(match.Prize.AddedPoints), int32(match.Prize.AddedCoins))
		}
	}
}

func (w *WordXMM) ResolveRound(me uuid.UUID, matchId uuid.UUID, points int) (err error, m *MatchToClient) {
	err, m = w.GetMatchFor(matchId, me)
	if err != nil {
		return
	} else {
		if m.Status != MS_PROGRESS {
			return status.Error(codes.InvalidArgument, "Match not in progress!"), nil
		}

		if m.WaitingUserID != nil && *m.WaitingUserID == m.OpponentID {
			return status.Error(codes.InvalidArgument, "Not your turn!"), nil
		}

		m.NextTurn(me, points, w.wx)
		err = w.SaveMatch(m)
		return
	}
}

func (w *WordXMM) RemoveMatch(matchId uuid.UUID) (err error) {
	_, err = w.db.Exec("DELETE from matches where id = $1", &matchId)
	return
}

func (w *WordXMM) SaveMatch(m *MatchToClient) (err error) {
	_, err = w.db.Exec("UPDATE matches SET (waitingUser, round, status, last_turn, user1_score, user2_score, win_user) = ($2, $3, $4, $5, $6, $7, $8) WHERE id = $1",
		&m.Id,
		m.WaitingUserID,
		&m.Round,
		&m.InnerStatus,
		&m.LastTurnTime,
		pq.Array(m.User1Score),
		pq.Array(m.User2Score),
		&m.WinUser,
	)

	return
}

func (w *WordXMM) GetMatchFor(matchId uuid.UUID, me uuid.UUID) (err error, match *MatchToClient) {
	rows, err := w.db.Query("SELECT id, user1, user2, waitingUser, round, status, invite_id, last_turn, user1_score, user2_score, win_user, words_seed from matches where id = $1", &matchId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "No such match"), nil
	}

	if rows.Next() {
		err, match = w.ScanMatch(rows, me, true)
	} else {
		return status.Error(codes.InvalidArgument, "No such match"), nil
	}

	if err != nil {
		return
	}

	return
}

func (w *WordXMM) StartMatch(matchId uuid.UUID, userID uuid.UUID, match *MatchToClient) (err error) {
	match.InnerStatus = MS_PROGRESS
	match.LastTurnTime = time.Now()
	match.Round = 1
	match.WaitingUserID = nil
	match.Update(userID, w.wx)
	err = w.SaveMatch(match)
	if err != nil {
		return
	}

	return
}

func (w *WordXMM) ScanMatch(rows *sql.Rows, me uuid.UUID, doUpdate bool) (error, *MatchToClient) {
	var matchId, user1Id, user2Id, inviteId uuid.UUID
	var waitingUser, winUser *uuid.UUID
	var round int
	var lastTurn time.Time
	var status string
	var seed int64
	var user1Score, user2Score []int64
	err := rows.Scan(&matchId, &user1Id, &user2Id, &waitingUser, &round, &status, &inviteId, &lastTurn, pq.Array(&user1Score), pq.Array(&user2Score), &winUser, &seed)
	if err != nil {
		return err, nil
	} else {

		m := &MatchToClient{
			User1:         user1Id,
			User2:         user2Id,
			WinUser:       winUser,
			Id:            matchId,
			WaitingUserID: waitingUser,
			InviteId:      inviteId,
			Round:         round,
			InnerStatus:   status,
			LastTurnTime:  lastTurn,
			User1Score:    user1Score,
			User2Score:    user2Score,
			Seed:          seed,
		}

		if doUpdate {
			m.Update(me, w.wx)
			w.UpdateWords(m, m.Seed)
		}

		return nil, m
	}
}

func (w *WordXMM) GetMatchesForUser(user uuid.UUID, filterMatchId *uuid.UUID) (error, []*MatchToClient) {
	var rows *sql.Rows
	var err error
	if filterMatchId == nil {
		rows, err = w.db.Query("SELECT id, user1, user2, waitingUser, round, status, invite_id, last_turn, user1_score, user2_score, win_user, words_seed from matches where user1 = $1 OR user2 = $1", &user)
	} else {
		rows, err = w.db.Query("SELECT id, user1, user2, waitingUser, round, status, invite_id, last_turn, user1_score, user2_score, win_user, words_seed from matches where id = $1", filterMatchId)
	}

	if err != nil {
		return err, nil
	}

	var matches []*MatchToClient
	for rows.Next() {
		err, match := w.ScanMatch(rows, user, true)

		if err != nil {
			return err, nil
		}
		if match.User1 != user && match.User2 != user {
			return status.Error(codes.InvalidArgument, "No permission to access this match"), nil
		}

		matches = append(matches, match)
	}

	return nil, matches
}

func (w *WordXMM) AddMatchFor(inviteId uuid.UUID, me uuid.UUID, id2 uuid.UUID) (err error, match *MatchToClient) {
	rows, err := w.db.Query("SELECT id from matches where ((user1 = $1 AND user2 = $2) OR (user2 = $1 AND user1 = $2)) AND status <> 'finished'", &me, &id2)
	if err != nil {
		return err, nil
	}
	alreadyHave := false

	if rows.Next() {
		alreadyHave = true
	}
	if alreadyHave {
		return status.Error(codes.InvalidArgument, "Already have a match"), nil
	} else {
		id := uuid.Must(uuid.NewV4())
		seed := rand.Intn(math.MaxInt32)
		invited := MS_INVITED
		round := 1
		_, err := w.db.Exec("INSERT into matches (id, user1, user2,  round, status, invite_id, last_turn, words_seed) VALUES ($1, $2, $3, $4, $5, $6, now(), $7)",
			&id,
			&me,
			&id2,
			&round,
			&invited,
			&inviteId,
			&seed,
		)
		if err != nil {
			return err, nil
		}

		m := &MatchToClient{
			Id:     id,
			User1:  me,
			User2:  id2,
			Status: invited,
			Seed:   int64(seed),
		}

		m.Update(me, w.wx)
		return nil, m
	}

}

type WordXMM struct {
	db   *sql.DB
	wx   *WordX
	rand *random.Random
	//	H2HMatches    map[uuid.UUID]MatchList //match by user id
	//	MatchesList   map[uuid.UUID]*Match // match by match id
	mu *sync.Mutex
}

func CreateWordXMM(db *sql.DB) *WordXMM {
	return &WordXMM{
		db:   db,
		rand: random.New(),
		//	H2HMatches: make(map[uuid.UUID]MatchList),
		//	MatchesList: make(map[uuid.UUID]*Match),
		mu: &sync.Mutex{},
	}
}
