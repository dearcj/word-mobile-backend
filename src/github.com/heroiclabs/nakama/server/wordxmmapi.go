package server

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/heroiclabs/nakama/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
	"time"
)

type Opponent struct {
	Coins       int
	Points      int
	Name        string
	Friend      bool
	Id          uuid.UUID
	Facebook_Id *string
}

func (s *ApiServer) GetMatches(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			MatchID string `json:"match_id,omitempty"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}
	var filter *uuid.UUID
	if b.Payload.MatchID != "" {
		matchid, err := uuid.FromString(b.Payload.MatchID)
		filter = &matchid
		if err != nil {
			return nil, err
		}

	}

	res := struct {
		Matches []*MatchToClient
	}{}

	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)
	err, res.Matches = s.runtimePool.wx.MM.GetMatchesForUser(userID, filter)
	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func ParamStr(numParams int, offset int) string {
	var paramCounters []string

	for inx := 0; inx < numParams; inx++ {
		paramCounters = append(paramCounters, "$"+strconv.Itoa(inx+1+offset))
	}

	return strings.Join(paramCounters, ",")
}

func (s *ApiServer) RequestUserInfos(ids []uuid.UUID) (error, map[uuid.UUID]*Opponent) {
	query := "SELECT id, display_name, facebook_id FROM users where id IN (" + ParamStr(len(ids), 0) + ")"
	var params []interface{}
	for _, x := range ids {
		idx := x
		params = append(params, &idx)
	}

	rows, err := s.db.Query(query, params...)
	if err != nil {
		return err, nil
	}
	var res = make(map[uuid.UUID]*Opponent)

	for rows.Next() {
		var id *uuid.UUID
		var dname *string
		var facebook_id *string
		err = rows.Scan(&id, &dname, &facebook_id)

		if err != nil {
			return err, nil
		}

		res[*id] = &Opponent{
			Id: *id,
		}

		res[*id].Facebook_Id = facebook_id

		if dname != nil {
			res[*id].Name = *dname
		}
	}

	var objs []*api.ReadStorageObjectId
	for k, _ := range res {
		objs = append(objs, &api.ReadStorageObjectId{
			UserId:     k.String(),
			Key:        "money",
			Collection: "resources",
		}, &api.ReadStorageObjectId{
			UserId:     k.String(),
			Key:        "points",
			Collection: "resources",
		})
	}

	storage, err := StorageReadObjects(s.runtimePool.logger, s.db, uuid.Must(uuid.FromString("00000000-0000-0000-0000-000000000000")), objs)
	for _, x := range storage.Objects {
		_, vstr := StorageValueGet(x.Value)
		v, err := strconv.Atoi(vstr)
		if err != nil {
			return err, nil
		}

		id, err := uuid.FromString(x.UserId)
		if err != nil {
			return err, nil
		}
		if x.Key == "money" {
			res[id].Coins = v
		}

		if x.Key == "points" {
			res[id].Points = v
		}
	}

	if err != nil {
		return err, nil
	}

	return nil, res
}

func (s *ApiServer) GetUserFriends(userID uuid.UUID, limit int) (error, []uuid.UUID) {
	var res []uuid.UUID

	query := "SELECT destination_id FROM user_edge where source_id = $1 and state = 0 LIMIT $2"
	rows, err := s.db.Query(query, &userID, &limit)
	if err != nil {
		return err, nil
	}

	for rows.Next() {
		var id uuid.UUID
		rows.Scan(&id)

		res = append(res, id)
	}

	return nil, res
}

func (s *ApiServer) FindUserByName(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Name string `json:"name"`
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	name := strings.ToLower(b.Name)
	query := "SELECT id FROM users where LOWER(display_name) = $1 "
	rows, err := s.db.Query(query, &name)

	if err != nil {
		return nil, err
	}

	var ids []uuid.UUID
	for rows.Next() {
		var id *uuid.UUID

		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		ids = append(ids, *id)
	}

	if len(ids) == 0 {
		return &api.WordXAPIRes{Payload: "{}"}, nil
	}

	err, users := s.RequestUserInfos(ids)

	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(&users)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) GetOpponents(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	const TOTAL_OPP = 20

	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	b := struct {
		Opponents  []*Opponent
		Matches    []*MatchToClient
		FriendsNum int
	}{}

	var err error
	err, b.Matches = s.runtimePool.wx.MM.GetMatchesForUser(userID, nil)

	if err != nil {
		return nil, err
	}

	err, friends := s.GetUserFriends(userID, 10)

	if err != nil {
		return nil, err
	}

	for _, f := range friends {
		b.Opponents = append(b.Opponents, &Opponent{
			Id:     f,
			Friend: true,
		})
	}

	//skipping me, skipping
	var params []interface{}
	for _, f := range friends {
		params = append(params, &f)
	}

	params = append(params, &userID, &uuid.Nil)

	limit := TOTAL_OPP + len(b.Matches)
	query := "SELECT id FROM users where id not in (" + ParamStr(len(params), 0) + ") ORDER BY last_login DESC LIMIT $" + strconv.Itoa(len(params)+1)
	params = append(params, &limit)
	rows, err := s.db.Query(query, params...)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		opp := &Opponent{}

		var id *uuid.UUID

		err := rows.Scan(&id)
		opp.Id = *id
		if err != nil {
			return nil, err
		}

		b.Opponents = append(b.Opponents, opp)
	}

	for opinx := 0; opinx < len(b.Opponents); opinx++ {
		found := false
		for _, x := range b.Matches {
			if b.Opponents[opinx].Id == x.OpponentID {
				found = true
			}
		}

		if found {
			b.Opponents = append(b.Opponents[:opinx], b.Opponents[opinx+1:]...)
			opinx--
		}
	}

	if len(b.Opponents) > TOTAL_OPP {
		b.Opponents = b.Opponents[:TOTAL_OPP]
	}

	var ids []uuid.UUID
	for _, o := range b.Opponents {
		ids = append(ids, o.Id)
	}
	for _, o := range b.Matches {
		ids = append(ids, o.OpponentID)
	}
	err, users := s.RequestUserInfos(ids)

	if err != nil {
		return nil, err
	}

	for inx, o := range b.Opponents {
		if b.Opponents[inx].Friend {
			users[o.Id].Friend = true
		}
		b.Opponents[inx] = users[o.Id]
	}

	for _, o := range b.Matches {
		o.Opponent = users[o.OpponentID]
	}

	b.FriendsNum = len(friends)

	bytea, err := json.Marshal(&b)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

type InviteReq struct {
	Payload *Invite `json:"payload"`
}

type Invite struct {
	OpponentId string `json:"opponent_id"`
}

func (s *ApiServer) SendNotification(from uuid.UUID, to uuid.UUID, content string, subject string, code int32) (*api.Notification, error) {
	notid := uuid.Must(uuid.NewV4())
	notifications := make(map[uuid.UUID][]*api.Notification)
	notification := &api.Notification{}
	notification.Id = notid.String()
	notification.SenderId = from.String()
	notification.Persistent = true
	notification.Code = code
	notification.Subject = subject
	notification.CreateTime = &timestamp.Timestamp{Seconds: time.Now().UTC().Unix()}
	notification.Content = content
	notifications[to] = []*api.Notification{notification}

	if err := NotificationSend(s.logger, s.db, s.router, notifications); err != nil {
		return nil, status.Error(codes.Internal, "Can't send notification")
	} else {
		return notification, nil
	}
}

func (s *ApiServer) GiveUp(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			MatchID string `json:"match_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	matchid, err := uuid.FromString(b.Payload.MatchID)
	if err != nil {
		return nil, err
	}
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	err, m := s.runtimePool.wx.MM.GetMatchFor(matchid, userID)

	if err != nil {
		return nil, err
	}

	if m.InnerStatus != MS_PROGRESS {
		return nil, status.Error(codes.InvalidArgument, "Can't give up. Match not in progress")
	}

	m.WinUser = &m.OpponentID
	m.InnerStatus = MS_FINISHED
	m.Update(userID, s.runtimePool.wx)
	s.runtimePool.wx.MM.SaveMatch(m)

	res := struct {
		Message string
		Match   *MatchToClient
	}{
		Message: "Gave up successfully",
		Match:   m,
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}

	err = s.NotifyYourTurn(m, userID, 1)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) ResolveH2HRound(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			AddedPoints float32 `json:"added_points"`
			MatchID     string  `json:"match_id"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	matchid, err := uuid.FromString(b.Payload.MatchID)
	if err != nil {
		return nil, err
	}
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	err, m := s.runtimePool.wx.MM.ResolveRound(userID, matchid, int(b.Payload.AddedPoints))
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, status.Error(codes.InvalidArgument, "Can't resolve round")
	}

	res := struct {
		Message string
		Match   *MatchToClient
	}{
		Message: "Round resolved",
		Match:   m,
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, err
	}

	err = s.NotifyYourTurn(m, userID, 1)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) ResolveInvite(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			MatchID    string `json:"match_id"`
			Resolution string `json:"resolution"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	lower := strings.ToLower(b.Payload.Resolution)

	matchID, err := uuid.FromString(b.Payload.MatchID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid match_id param")
	}
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	err, match := s.runtimePool.wx.MM.GetMatchFor(matchID, userID)
	if err != nil {
		return nil, err
	}
	if match.Status != MS_INVITED {
		return nil, status.Error(codes.InvalidArgument, "Invite already accepted")
	}

	if match.User1 != userID && match.User2 != userID {
		return nil, status.Error(codes.InvalidArgument, "Not your invite")
	}

	//if match.InviteId.String() != b.Payload.InviteID {
	//	return nil, status.Error(codes.InvalidArgument, "incorrect invite_id")
	//}

	if lower == "accept" {
		return s.AcceptInvite(ctx, matchID, userID, match)
	}

	if lower == "decline" {
		return s.DeclineInvite(ctx, matchID, userID, match)
	}

	return nil, status.Error(codes.InvalidArgument, "Invalid resolution param, accepts only ['accept' or 'decline']")
}

func (s *ApiServer) AcceptInvite(ctx context.Context, matchID uuid.UUID, userID uuid.UUID, match *MatchToClient) (*api.WordXAPIRes, error) {
	err := s.runtimePool.wx.MM.StartMatch(matchID, userID, match)
	if err != nil {
		return nil, err
	}

	b := struct {
		Message string
	}{
		Message: "Invite accepted",
	}

	bytea, err := json.Marshal(&b)
	if err != nil {
		return nil, err
	}

	dest := match.OpponentID
	_, err = s.SendNotification(userID, dest, match.toString(dest, s.runtimePool.wx), "invite accepted", 1)

	if err != nil {
		return nil, err
	}

	s.runtimePool.wx.SendPushTo(dest, "Your invite accepted", "")

	err = s.NotifyYourTurn(match, userID, 1)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) NotifyYourTurn(m *MatchToClient, me uuid.UUID, code int32) error {
	var zeroUUID, dest uuid.UUID
	if m.Status == MS_FINISHED {
		if m.WinUser == nil {
			to := m.User1

			_, err := s.SendNotification(zeroUUID, to, m.toString(m.User1, s.runtimePool.wx), "Draw", code)

			s.runtimePool.wx.SendPushTo(to, "Match ended in a draw", "")

			if err != nil {
				return err
			}

			to = m.User2
			_, err = s.SendNotification(zeroUUID, to, m.toString(m.User2, s.runtimePool.wx), "Draw", code)
			s.runtimePool.wx.SendPushTo(to, "Match ended in a draw", "")

			if err != nil {
				return err
			}

		} else {
			var winner, loser uuid.UUID
			if *m.WinUser == m.User1 {
				winner = m.User1
				loser = m.User2
			} else {
				winner = m.User2
				loser = m.User1
			}
			_, err := s.SendNotification(zeroUUID, winner, m.toString(winner, s.runtimePool.wx), "Won fight", code)
			s.runtimePool.wx.SendPushTo(winner, "Congratulations! You won a match", "")

			if err != nil {
				return err
			}

			_, err = s.SendNotification(zeroUUID, loser, m.toString(loser, s.runtimePool.wx), "Lose fight", code)
			s.runtimePool.wx.SendPushTo(loser, "Match lost, good luck next time", "")

			if err != nil {
				return err
			}
		}
		return nil

	} else {
		if m.WaitingUserID == nil {
			to := m.User1
			_, err := s.SendNotification(zeroUUID, to, m.toString(m.User1, s.runtimePool.wx), "Waiting both players", code)
			s.runtimePool.wx.SendPushTo(to, "Waiting for your turn", "")

			if err != nil {
				return err
			}

			to = m.User2
			_, err = s.SendNotification(zeroUUID, to, m.toString(m.User2, s.runtimePool.wx), "Waiting both players", code)
			s.runtimePool.wx.SendPushTo(to, "Waiting for your turn", "")

			if err != nil {
				return err
			}

			return nil
		} else {
			if *m.WaitingUserID == me {
				dest = me
			}

			if *m.WaitingUserID == m.OpponentID {
				dest = m.OpponentID
			}

			s.runtimePool.wx.SendPushTo(dest, "Waiting for your turn", "")
			_, err := s.SendNotification(zeroUUID, dest, m.toString(dest, s.runtimePool.wx), "Waiting you", code)
			if err != nil {
				return err
			} else {

				return nil
			}

		}
	}
}

func (s *ApiServer) DeclineInvite(ctx context.Context, matchID uuid.UUID, userID uuid.UUID, match *MatchToClient) (*api.WordXAPIRes, error) {
	err := s.runtimePool.wx.MM.RemoveMatch(matchID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Can't remove invitation")
	} else {
		results := struct {
			Msg string
		}{
			Msg: "Successfully declined",
		}

		userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

		bytea, err := json.Marshal(&results)
		if err != nil {
			return nil, err
		}

		_, err = s.SendNotification(userID, match.OpponentID, match.toString(match.OpponentID, s.runtimePool.wx), "invite declined", 1)
		if err != nil {
			return nil, err
		}

		s.runtimePool.wx.SendPushTo(match.OpponentID, "Your invitation is declined", "")

		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) SendInvite(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var invreq InviteReq
	err := json.Unmarshal([]byte(in.Payload), &invreq)
	if err != nil {
		return nil, err
	}
	if invreq.Payload == nil || invreq.Payload.OpponentId == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid OpponentId")
	}

	destId, err := uuid.FromString(invreq.Payload.OpponentId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid OpponentId")
	}

	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	notid := uuid.Must(uuid.NewV4())

	err, m := s.runtimePool.wx.MM.AddMatchFor(notid, userID, destId)
	if err != nil {
		return nil, err
	}

	b := struct {
		Match *MatchToClient
	}{
		Match: m,
	}

	bytea, err := json.Marshal(&b)
	if err != nil {
		return nil, err
	}

	_, err = s.SendNotification(userID, destId, m.toString(m.OpponentID, s.runtimePool.wx), "H2HInvite", 1)
	s.runtimePool.wx.SendPushTo(destId, "Invitation into head to head match", "")

	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}
