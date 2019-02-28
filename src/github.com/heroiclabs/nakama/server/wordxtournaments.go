package server

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama/api"
	//"github.com/golang/protobuf/ptypes/timestamp"
	"database/sql"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TourUser struct {
	Id           uuid.UUID  `json:"id"`
	TournamentId uuid.UUID  `json:"tournament_id"`
	Points       int        `json:"points"`
	TourID       *uuid.UUID `json:"tour_id"`
	UserID       uuid.UUID  `json:"user_id"`
	Submitted    bool       `json:"submitted"`
	Status       int        `json:"status"`
	Award        int        `json:"award"`
	StatusStr    string     `json:"status_str"`
}

type Tour struct {
	Id           *uuid.UUID `json:"id"`
	TournamentId uuid.UUID  `json:"tournament_id"`
	EndDate      time.Time  `json:"end_date"`
	Award        int32      `json:"award"`
	TopPerc      int32      `json:"top_perc"`
}

type ScoreUser struct {
	Id          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Place       int       `json:"place"`
	Status      string    `json:"status"`
	Winner      bool      `json:"winner"`
	Points      int       `json:"points"`
	Award       int       `json:"award"`
	Facebook_Id *string   `json:"facebook_id"`
}

type ScoresTable []*ScoreUser

type Tournament struct {
	Id              uuid.UUID    `json:"id"`
	Name            string       `json:"name"`
	ImageData       string       `json:"image_data"`
	ImageDataURL    string       `json:"image_data_url"`
	Language        int          `json:"language"`
	Location        string       `json:"location"`
	StartDate       *time.Time   `json:"start_date"`
	Participants    int32        `json:"participants"`
	CurParticipants int32        `json:"cur_participants"`
	Status          int32        `json:"status"`
	CurrentRoundID  *uuid.UUID   `json:"current_round_id"`
	CurrentRound    *Tour        `json:"current_round"`
	CurrentRoundNum int32        `json:"current_round_num"`
	Rounds          []*Tour      `json:"rounds"`
	StatusStr       string       `json:"status_str"`
	LanguageStr     string       `json:"language_str"`
	Desc            string       `json:"desc"`
	TourUser        *TourUser    `json:"tour_user"`
	ScoresTable     *ScoresTable `json:"scores_table"`
}

const (
	TS_UNPUB = 0
	TS_PUB   = 1
	TS_ACT   = 2
	TS_FIN   = 3
)

const (
	TUS_NORMAL = 0
	TUS_LEFT   = 1
)

func (t *Tournament) GetNextRound() (*Tour, error) {
	if t.CurrentRoundID == nil {
		if len(t.Rounds) == 0 {
			return nil, status.Error(codes.Internal, "NO ROUNDS!")
		}
		return t.Rounds[0], nil
	} else {
		for _, v := range t.Rounds {
			if v.EndDate.After(t.CurrentRound.EndDate) {
				return v, nil
			}
		}
		return nil, nil
	}
}

var tournamentStatus = map[int]string{
	0: "unpublished",
	1: "published",
	2: "active",
	3: "finished",
}

func (s *ApiServer) MassNotify(users []uuid.UUID, msg interface{}, code int32, subject string) error {
	bytea, _ := json.Marshal(msg)
	msgstr := string(bytea)

	from := uuid.FromBytesOrNil(nil)
	notifications := make(map[uuid.UUID][]*api.Notification)
	for _, u := range users {
		notification := &api.Notification{}
		notification.Id = uuid.Must(uuid.NewV4()).String()
		notification.SenderId = from.String()
		notification.Persistent = true
		notification.Code = code
		notification.Subject = subject
		notification.CreateTime = &timestamp.Timestamp{Seconds: time.Now().UTC().Unix()}
		notification.Content = msgstr
		notifications[u] = []*api.Notification{notification}
		//		s.SendNotification(from, u, msgstr, subject, 1)
	}

	if err := NotificationSend(s.logger, s.db, s.router, notifications); err != nil {
		return status.Error(codes.Internal, "Can't send notification")
	} else {
		return nil
	}
}

func ScanTourUser(rows *sql.Rows) (*TourUser, error) {
	var t TourUser

	err := rows.Scan(&t.Id,
		&t.Points,
		&t.Submitted,
		&t.TourID,
		&t.UserID,
		&t.TournamentId,
		&t.Status,
		&t.Award)

	return &t, err
}

func (s *ApiServer) GetTourUsers(rid *uuid.UUID, tid uuid.UUID) ([]*TourUser, error) {
	rows, err := s.db.Query("select id,  points, submitted, tour_id, user_id, tournament_id, status, award_money from tour_user where tournament_id = $1 and ((tour_id = $2) or (tour_id is null and $2 is null))", &tid, &rid)
	if err != nil {
		return nil, err
	}

	var tr []*TourUser

	for rows.Next() {

		t, err := ScanTourUser(rows)

		if err != nil {
			return nil, err
		}

		tr = append(tr, t)
	}

	sort.Slice(tr, func(i, j int) bool { return tr[i].Points > tr[j].Points })

	return tr, nil
}

func (s *ApiServer) MoveUsersToRound(tid uuid.UUID, users []uuid.UUID, tour *uuid.UUID) error {
	var params []interface{}
	var par []string
	var count = 1
	for _, u := range users {
		newid := uuid.Must(uuid.NewV4())
		uu := u
		params = append(params,
			&newid,
			&tid,
			&uu,
			&tour,
		)
		str := ""
		str += "$" + strconv.Itoa(count) + ", "
		str += "$" + strconv.Itoa(count+1) + ", "
		str += "$" + strconv.Itoa(count+2) + ", "
		str += "$" + strconv.Itoa(count+3) + ""

		par = append(par, "( "+str+" )")
		count += 4
	}
	if len(users) > 0 {
		_, err := s.db.Exec("INSERT INTO tour_user (id, tournament_id, user_id, tour_id) VALUES "+strings.Join(par, ","), params...)
		return err
	} else {
		return nil
	}
}

func (s *ApiServer) setRound(r *Tour, t *Tournament) {
	prev := t.CurrentRound
	next := r

	msg := struct {
		TournamentID uuid.UUID
	}{
		TournamentID: t.Id,
	}

	if prev == nil {
		t.Status = TS_ACT
		t.CurrentRoundNum = 1
		t.CurrentRoundID = next.Id

		users, err := s.GetTourUsers(nil, t.Id)
		if err != nil {
			s.logger.Error("can't get tour users", zap.Error(err))
			return
		}

		var userIDs []uuid.UUID
		for _, u := range users {
			userIDs = append(userIDs, u.UserID)
		}

		err = s.MassNotify(userIDs, &msg, 1, "Tournament is active")

		s.runtimePool.wx.SendMassPushTo(userIDs, "Tournament just started", "")

		if err != nil {
			s.logger.Error("Can't notify active tournament", zap.Error(err))
			return
		}

		err = s.MoveUsersToRound(t.Id, userIDs, next.Id)
		if err != nil {
			s.logger.Error("Can't move to next round", zap.Error(err))
			return
		}

		err = s.updateTournament(t, false)
		if err != nil {
			s.logger.Error("can't update tournament", zap.Error(err))
			return
		}
		//from nil round move all players to first round
	} else {
		users, err := s.GetTourUsers(prev.Id, t.Id)
		if err != nil {
			s.logger.Error("can't get tour users", zap.Error(err))
			return
		}

		numUsersToNextRound := int(float32(len(users))*(float32(prev.TopPerc)/100.) + 1)

		var winners, losers []*TourUser
		var winIDs, loseIDs []uuid.UUID

		for inx, u := range users {
			if inx < numUsersToNextRound {
				winners = append(winners, u)
				winIDs = append(winIDs, u.UserID)
			} else {
				losers = append(losers, u)
				loseIDs = append(loseIDs, u.UserID)
			}
		}

		for _, x := range winners {
			x.Award = int(t.CurrentRound.Award)
			s.AddResource(x.UserID, 0, int32(float32(t.CurrentRound.Award)))
		}

		err = s.UpdateTourUsers(winners)
		if err != nil {
			s.logger.Error("can't update tour_users", zap.Error(err))
			return
		}

		if next == nil {
			//	t.CurrentRoundID = nil
			t.Status = TS_FIN

			s.runtimePool.wx.SendMassPushTo(winIDs, "Congratulations!", " You won tournament")
			err = s.MassNotify(winIDs, &msg, 1, "Won tournament")
			if err != nil {
				s.logger.Error("Can't notify active tournament", zap.Error(err))
				return
			}

			err = s.MassNotify(loseIDs, &msg, 1, "Tournament flew out")
			if err != nil {
				s.logger.Error("Can't notify active tournament", zap.Error(err))
				return
			}

		} else {
			t.CurrentRoundID = next.Id
			t.CurrentRoundNum++

			err = s.MassNotify(winIDs, &msg, 1, "Tournament won round")
			if err != nil {
				s.logger.Error("Can't notify active tournament", zap.Error(err))
				return
			}

			err = s.MassNotify(loseIDs, &msg, 1, "Tournament flew out")
			if err != nil {
				s.logger.Error("Can't notify active tournament", zap.Error(err))
				return
			}

			err = s.MoveUsersToRound(t.Id, winIDs, next.Id)
			if err != nil {
				s.logger.Error("Can't notify active tournament", zap.Error(err))
				return
			}
		}

		err = s.updateTournament(t, false)
		if err != nil {
			s.logger.Error("can't update tournament", zap.Error(err))
			return
		}
	}
}

func (s *ApiServer) DashboardAddTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	uid := uuid.Must(uuid.NewV4())

	DEF_PARTICIPANTS := 100
	query := "INSERT INTO tournament (id, participants) VALUES ($1, $2);"
	_, err := s.db.Exec(query, &uid, &DEF_PARTICIPANTS)

	if err != nil {
		return nil, err
	}

	results := &Tournament{
		Id:           uid,
		ImageDataURL: "/images/" + uid.String(),
	}

	bytea, err := json.Marshal(&results)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func ScanTour(r *sql.Rows) (tour *Tour, err error) {
	tour = &Tour{}
	err = r.Scan(&tour.Id, &tour.TournamentId, &tour.EndDate, &tour.Award, &tour.TopPerc)
	return
}

func ScanTournament(r *sql.Rows, scanImages bool) (*Tournament, error) {
	t := &Tournament{}
	var par *int32

	var str sql.NullString
	args := []interface{}{
		&t.Id,
		&t.Name,
		&t.Location,
		&t.StartDate,
		&t.Participants,
		&t.Status,
		&t.CurrentRoundID,
		&t.CurrentRoundNum,
		&t.Language,
		&t.Desc,
		&par,
	}

	if scanImages == true {
		args = append(args, &str)
	}

	err := r.Scan(args...)

	if par != nil {
		t.CurParticipants = *par
	}

	if err != nil {
		return nil, err

	}

	t.ImageDataURL = "/image/" + t.Id.String()

	if str.Valid {
		t.ImageData = str.String
	}

	t.StatusStr = tournamentStatus[int(t.Status)]
	t.LanguageStr = WX_LANGS[t.Language]

	return t, nil
}

func (s *ApiServer) UpdateTourUsers(tu []*TourUser) error {
	for _, t := range tu {
		_, err := s.db.Exec("update tour_user set (submitted, points, award_money, status) = ($1, $2, $3, $4) where id = $5", &t.Submitted, &t.Points, &t.Award, &t.Status, &t.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ApiServer) TournamentMakeTurn(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	var b struct {
		Payload struct {
			TournamentId *uuid.UUID `json:"tournament_id"`
			Points       int        `json:"points"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if &b.Payload.TournamentId == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	t, err := s.GetTournament(*b.Payload.TournamentId, false)
	if err != nil {
		return nil, err
	}

	if t.CurrentRound == nil || t.Status != TS_ACT {
		return nil, StatusError(codes.Internal, "Tournament not in active status or current round is nil", nil)
	}

	s.AddResource(userID, int32(b.Payload.Points), 0)

	curround := *t.CurrentRound
	tr := true
	_, err = s.db.Exec("update tour_user set (submitted, points) = ($1, $2) where tour_id = $3 and user_id =$4", &tr, &b.Payload.Points, &curround.Id, &userID)
	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "Score submitted",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) LeaveTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	var b struct {
		Payload struct {
			TournamentId *uuid.UUID `json:"tournament_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if &b.Payload.TournamentId == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	t, err := s.GetTournament(*b.Payload.TournamentId, false)
	if err != nil {
		return nil, err
	}

	tt := []*Tournament{t}
	err = s.GetTournamentsUserData(userID, tt)
	if tt[0].TourUser == nil {
		return nil, StatusError(codes.Internal, "No tour user record", nil)
	}

	if t.Status == TS_PUB {
		_, err = s.db.Exec("delete from tour_user where id = $1", &t.TourUser.Id)
	} else {
		left := TUS_LEFT
		_, err = s.db.Exec("update tour_user set (status) = ($1) where user_id = $2 and tournament_id = $3", &left, &userID, t.Id)
	}

	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "Left tournament",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) JoinTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			TournamentId *uuid.UUID `json:"tournament_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	ids, err := s.GetUserTournaments(userID)
	if err != nil {
		return nil, err
	}

	if len(ids) > 0 {
		return nil, StatusError(codes.Internal, "You already joined one tournament", nil)
	}

	joined, err := s.GetUserTournaments(userID)
	if err != nil {
		return nil, err
	}

	if len(joined) > 0 {
		return nil, StatusError(codes.Internal, "You already joined one tournament", nil)
	}

	if &b.Payload.TournamentId == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	t, err := s.GetTournament(*b.Payload.TournamentId, false)

	if err != nil {
		return nil, err
	}

	if t.CurParticipants >= t.Participants {
		return nil, StatusError(codes.Internal, "Tournament is full", nil)
	}

	if t.Status != TS_PUB {
		return nil, StatusError(codes.Internal, "Can't join non-active tournament", nil)
	}

	newid := uuid.Must(uuid.NewV4())

	_, err = s.db.Exec("INSERT INTO tour_user (id, tournament_id, user_id) VALUES ($1, $2, $3)", &newid, &t.Id, &userID)

	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "Joined tournament",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) GetAvailableTournaments(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	q := "SELECT location FROM users where id = $1"

	var location string

	err := s.db.QueryRow(q, &userID).Scan(&location)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(TournamentQuery("", " where tt.status = 1 and (tt.location = 'any' or tt.location = '"+location+"') "))
	if err != nil {
		return nil, err
	}

	var tournaments = make(map[uuid.UUID]*Tournament)

	for rows.Next() {
		t, err := ScanTournament(rows, false)
		if err != nil {
			return nil, err
		}

		tournaments[t.Id] = t
	}

	cur, err := s.GetUserTournaments(userID)
	if err != nil {
		return nil, err
	}

	for _, tt := range cur {
		delete(tournaments, *tt)
	}

	var tSlice []*Tournament
	for _, v := range tournaments {
		tSlice = append(tSlice, v)
	}

	err = s.GetTournamentsRounds(tSlice)
	if err != nil {
		return nil, err
	}

	res := struct {
		Tournaments []*Tournament `json:"tournaments"`
	}{
		Tournaments: tSlice,
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) GetUserTournaments(user uuid.UUID) ([]*uuid.UUID, error) {
	rows, err := s.db.Query("select tu.tournament_id from tour_user tu, tournament t where tu.status = 0 and t.id = tu.tournament_id and (t.status = 1 or t.status = 2 or t.status = 3) and tu.user_id = $1 and ((tu.tour_id = t.current_round) or (tu.tour_id is null and t.current_round is null)) group by tournament_id", &user)

	if err != nil {
		return nil, err
	}

	var tournaments []*uuid.UUID
	for rows.Next() {
		var t uuid.UUID
		err := rows.Scan(&t)

		if err != nil {
			return nil, err
		}

		tournaments = append(tournaments, &t)
	}

	return tournaments, nil
}

func (s *ApiServer) InnerFilterTournaments(cond string, userID uuid.UUID) ([]*Tournament, error) {
	var list []*Tournament
	tors, err := s.GetUserTournaments(userID)
	if err != nil {
		return nil, err
	}

	var torInterfaces []interface{}
	for _, x := range tors {
		torInterfaces = append(torInterfaces, x)
	}

	filter := " where " + strings.Join([]string{cond, " tt.id IN ( " + ParamStr(len(torInterfaces), 0) + " )"}, " AND ")
	if len(tors) > 0 {
		rows, err := s.db.Query(TournamentQuery("", filter), torInterfaces...)

		if err != nil {
			return nil, err
		}

		for rows.Next() {
			t, err := ScanTournament(rows, false)

			if err != nil {
				return nil, err
			}

			list = append(list, t)
		}
	}

	return list, nil
}

func (s *ApiServer) GetScoreTable(tid uuid.UUID, curRound *uuid.UUID) (*ScoresTable, error) {
	st := ScoresTable{}

	uu, e := s.GetTourUsers(curRound, tid)
	if e != nil {
		return nil, e
	}

	inx := 0
	var queries []string
	var params []interface{}
	for _, u := range uu {
		status := "inprogress"

		if u.Status == 1 {
			status = "left"
		}

		var winner = false
		if u.Award > 0 {
			winner = true
		}

		if u.Award < 0 {
			u.Award = 0
		}

		st = append(st, &ScoreUser{
			Id:          u.UserID,
			Status:      status,
			DisplayName: "",
			Place:       inx + 1,
			Winner:      winner,
			Award:       u.Award,
			Points:      u.Points,
		})

		params = append(params, &u.UserID)
		queries = append(queries, "SELECT display_name, facebook_id FROM users where id = $"+strconv.Itoa(inx+1))
		inx++
	}
	query := strings.Join(queries, " UNION ")

	rows, e := s.db.Query(query, params...)

	if e != nil {
		return nil, e
	}

	inx = 0
	for rows.Next() {
		var dname string
		var fbid *string
		rows.Scan(&dname, &fbid)

		if inx < len(st) {
			st[inx].Facebook_Id = fbid
			st[inx].DisplayName = dname
		}

		inx++
	}

	return &st, nil
}

func (s *ApiServer) GetActiveTournaments(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	list, err := s.InnerFilterTournaments("tt.status = 2 or tt.status = 1 or tt.status = 3", userID)
	if err != nil {
		return nil, err
	}

	mp := make(map[uuid.UUID]*Tournament)
	for _, x := range list {
		t := x
		mp[x.Id] = t
	}

	cur, err := s.GetUserTournaments(userID)
	if err != nil {
		return nil, err
	}

	for _, x := range mp {
		found := false
		for _, y := range cur {
			if *y == x.Id {
				found = true
			}
		}

		if !found {
			delete(mp, x.Id)
		}
	}

	list = nil
	for _, x := range mp {
		list = append(list, x)
	}

	if len(list) > 0 {
		err = s.GetTournamentsRounds(list)
		if err != nil {
			return nil, err
		}

		err = s.GetTournamentsUserData(userID, list)
		if err != nil {
			return nil, err
		}
	}

	if len(list) > 0 {
		t := list[0]
		t.ScoresTable, err = s.GetScoreTable(t.Id, t.CurrentRoundID)

		if err != nil {
			return nil, err
		}
	}

	res := struct {
		Tournaments []*Tournament `json:"tournaments"`
	}{
		Tournaments: list,
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) FinishTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	return nil, StatusError(codes.Internal, "Deprecated method", nil)
}

func (s *ApiServer) DashboardPublishTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			TournamentId *uuid.UUID `json:"tournament_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if &b.Payload.TournamentId == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	t, err := s.GetTournament(*b.Payload.TournamentId, false)
	if err != nil {
		return nil, err
	}

	if len(t.Rounds) == 0 {
		return nil, StatusError(codes.Internal, "Can't publish tournament with no rounds", nil)
	}

	t.Status = TS_PUB
	err = s.updateTournament(t, false)
	if err != nil {
		return nil, err
	}

	res := struct {
		Message    string
		Tournament *Tournament `json:"tournament"`
	}{
		Message:    "Tournament published",
		Tournament: t,
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func TournamentQuery(imgDataParamStr string, whereFilter string) string {
	query :=
		"select tt.id, tt.name, tt.location, tt.start_date, tt.participants, tt.status, tt.current_round, tt.current_round_num, tt.lang, tt.description, e.part_cur " + imgDataParamStr + " from  (select t.id as tid, count(tu.id) as part_cur from tournament t join tour_user tu on t.id = tu.tournament_id and ((t.current_round = tu.tour_id) or (t.current_round is null and tu.tour_id is null)) group by t.id) e right join tournament tt on tt.id = e.tid " + whereFilter

	return query
}

func (s *ApiServer) GetTournament(tid uuid.UUID, scanImages bool) (*Tournament, error) {
	var imdataStr string

	if scanImages {
		imdataStr = ", t.image_data "
	}

	filter := " where tt.id = $1 "

	tq := TournamentQuery(imdataStr, filter)
	rows, err := s.db.Query(tq, &tid)

	if err != nil {
		return nil, err
	}

	var t *Tournament
	for rows.Next() {
		t, err = ScanTournament(rows, scanImages)

		if err != nil {
			return nil, err
		}
	}

	res := []*Tournament{t}
	err = s.GetTournamentsRounds(res)
	if err != nil {
		return nil, err
	}

	return res[0], nil
}

func (s *ApiServer) updateTournament(t *Tournament, updateImage bool) error {
	var imdata string
	var imparam string
	if updateImage {
		imdata = ", image_data "
		imparam = ", $11"
	}
	query := "UPDATE tournament SET (name, location, start_date, participants, status, lang, description, current_round, current_round_num" + imdata + ") = ($2, $3, $4, $5, $6, $7, $8, $9, $10 " + imparam + " ) where id = $1;"

	var params = []interface{}{
		&t.Id,
		&t.Name,
		&t.Location,
		t.StartDate,
		&t.Participants,
		&t.Status,
		&t.Language,
		&t.Desc,
		&t.CurrentRoundID,
		&t.CurrentRoundNum,
	}

	if updateImage {
		params = append(params, &t.ImageData)
	}

	_, err := s.db.Exec(query,
		params...,
	)

	return err
}

func (s *ApiServer) DashboardUpdateTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Tournament   *Tournament `json:"tournament"`
			DeleteRounds []uuid.UUID `json:"delete_rounds"`
		}
	}
	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if &b.Payload.Tournament == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	err = s.updateTournament(b.Payload.Tournament, true)

	if err != nil {
		return nil, err
	}

	var inx = 1
	var counter = func() string {

		i := strconv.Itoa(inx)
		inx++
		return "$" + i
	}

	var deleteParams []interface{}

	for _, x := range b.Payload.DeleteRounds {
		id := x
		deleteParams = append(deleteParams, &id)
	}

	if len(deleteParams) > 0 {
		_, err = s.db.Exec("delete from tour where id in ("+ParamStr(len(deleteParams), 0)+")", deleteParams...)

		if err != nil {
			return nil, err
		}
	}

	var params []interface{}
	var roundsUpdates []string
	for _, x := range b.Payload.Tournament.Rounds {
		if x.Id == nil {
			u := uuid.Must(uuid.NewV4())
			x.Id = &u
		}

		roundsUpdates = append(roundsUpdates, " ("+strings.Join([]string{counter(), counter(), counter(), counter(), counter()}, ",")+" )")
		params = append(params,
			x.Id,
			&x.TournamentId,
			&x.TopPerc,
			&x.EndDate,
			&x.Award,
		)
	}

	if len(params) > 0 {
		_, err = s.db.Exec("UPSERT INTO tour (id, tournament_id, top_perc, end_date, award) VALUES "+strings.Join(roundsUpdates, ", "), params...)
		if err != nil {
			return nil, err
		}
	}

	res := struct {
		Message    string
		Tournament *Tournament `json:"tournament"`
	}{
		Message:    "Tournament updated",
		Tournament: b.Payload.Tournament,
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardDeleteTour(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			TourID *uuid.UUID `json:"tour_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if &b.Payload.TourID == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	todelete := *b.Payload.TourID
	query := "delete from tour where id = $1;"
	_, err = s.db.Exec(query, &todelete)

	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "Tour deleted",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardDeleteTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			TournamentID *uuid.UUID `json:"tournament_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.TournamentID == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	todelete := *b.Payload.TournamentID
	query := "delete from tournament where id = $1;"
	_, err = s.db.Exec(query, &todelete)

	if err != nil {
		return nil, err
	}

	query = "delete from tour where tournament_id = $1;"
	_, err = s.db.Exec(query, &todelete)

	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "Tournament deleted",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) GetTournamentsInner(filters []bool, scanImage bool, page int) (int, []*Tournament, error) {
	filterCondStr := ""

	const TOURNAMENTS_PER_PAGE = 20
	if len(filters) == 4 {
		var filterCond []string
		for inx, x := range filters {
			if x {
				filterCond = append(filterCond, " tt.status = "+strconv.Itoa(inx))
			}
		}

		if len(filterCond) > 0 {
			filterCondStr = " where (" + strings.Join(filterCond, " OR ") + ") "
		}
	}

	filter := filterCondStr + " order by tt.create_date asc "
	var scanImageStr string
	if scanImage {
		scanImageStr = ", image_data "
	}

	var pages int
	err := s.db.QueryRow("select count(1) from tournament tt " + filterCondStr).Scan(&pages)
	if err != nil {
		return 0, nil, err
	}

	pages = int(math.Floor(float64(pages)/TOURNAMENTS_PER_PAGE)) + 1

	rows, err := s.db.Query(TournamentQuery(scanImageStr, filter))

	if err != nil {
		return 0, nil, err
	}

	var res []*Tournament
	for rows.Next() {
		t, err := ScanTournament(rows, scanImage)
		if err != nil {
			return 0, nil, err
		}
		res = append(res, t)
	}

	if len(res) > 0 {
		err := s.GetTournamentsRounds(res)
		if err != nil {
			return 0, nil, err
		}
	}

	return pages, res, nil
}

func (s *ApiServer) DashboardGetTournaments(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Filter []bool `json:"filter"`
			Page   int
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	pages, res, err := s.GetTournamentsInner(b.Payload.Filter, false, b.Payload.Page)

	if err != nil {
		return nil, err
	}

	results := struct {
		Tournaments []*Tournament
		Pages       int
	}{
		Tournaments: res,
		Pages:       pages,
	}

	bytea, err := json.Marshal(&results)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) GetTournamentsUserData(uid uuid.UUID, res []*Tournament) error {
	query := "select id,  points, submitted, tour_id, user_id, tournament_id, status, award_money  from tour_user where user_id = $1 and tournament_id IN (" + ParamStr(len(res)+1, 0) + ") "
	var params []interface{}
	params = append(params, &uid)
	for _, x := range res {
		idx := x.Id
		params = append(params, &idx)
	}

	rows, err := s.db.Query(query, params...)
	if err != nil {
		return err
	}

	for rows.Next() {
		t, err := ScanTourUser(rows)
		if err != nil {
			return err
		}

		for _, r := range res {
			if t.TournamentId == r.Id && ((t.TourID == nil && r.CurrentRoundID == nil) || (t.TourID != nil && r.CurrentRoundID != nil &&
				*t.TourID == *r.CurrentRoundID)) {
				r.TourUser = t
			}
		}
	}

	return nil
}

func (s *ApiServer) GetTournamentsRounds(res []*Tournament) error {
	if len(res) == 0 {
		return nil
	}

	query := "select id, tournament_id, end_date, award, top_perc from tour where tournament_id IN (" + ParamStr(len(res), 0) + ") order by end_date asc"
	var params []interface{}
	for _, x := range res {
		idx := x.Id
		params = append(params, &idx)
	}

	rows, err := s.db.Query(query, params...)
	if err != nil {
		return err
	}

	for rows.Next() {
		tour, err := ScanTour(rows)

		if err != nil {
			return err
		}

		for _, t := range res {
			if t.Id == tour.TournamentId {
				t.Rounds = append(t.Rounds, tour)
			}
		}

	}

	for _, v := range res {
		var currentRound *Tour
		for _, r := range v.Rounds {
			if v.CurrentRoundID != nil && r.Id != nil && *r.Id == *v.CurrentRoundID {
				currentRound = r
			}
		}
		v.CurrentRound = currentRound
	}

	return nil
}

func (s *ApiServer) DashboardGetTournament(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			TournamentID *uuid.UUID `json:"tournament_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.TournamentID == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	rows, err := s.db.Query(TournamentQuery(", image_data", " where tt.id = $1"), b.Payload.TournamentID)

	if err != nil {
		return nil, err
	}

	var res []*Tournament
	for rows.Next() {
		t, err := ScanTournament(rows, true)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}

	results := struct {
		Tournaments []*Tournament
	}{
		Tournaments: res,
	}

	if len(res) > 0 {
		s.GetTournamentsRounds(res)
	}

	bytea, err := json.Marshal(&results)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardTournamentForward(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			TournamentID *uuid.UUID `json:"tournament_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.TournamentID == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	rows, err := s.db.Query(TournamentQuery("", " where tt.id = $1"), b.Payload.TournamentID)

	if err != nil {
		return nil, err
	}

	var t *Tournament
	for rows.Next() {
		t, err = ScanTournament(rows, false)

		if err != nil {
			return nil, err
		}
	}

	if t == nil {
		return nil, StatusError(codes.Internal, "No such tournament", nil)
	}

	err = s.GetTournamentsRounds([]*Tournament{t})
	if err != nil {
		return nil, err
	}

	if t.Status == TS_PUB || t.Status == TS_ACT {
		s.runtimePool.wx.TournamentCron.cronTournament(t.Id, t)
	} else {
		return nil, StatusError(codes.Internal, "Tournament status not `active` and not `published`", nil)
	}

	res := struct {
		Tournament *Tournament
		Message    string
	}{
		Tournament: t,
		Message:    "Tournament moved forward",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}
