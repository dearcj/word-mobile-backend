package server

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/tbalthazar/onesignal-go"
	"net/http"
)

func playerIdFromJSON(jsonStr string) (error, string) {
	var b struct {
		Value string `json:"value"`
	}

	err := json.Unmarshal([]byte(jsonStr), &b)
	return err, b.Value
}

func (wx *WordX) SendMassPushTo(userID []uuid.UUID, subject string, content string) error {

	query := "SELECT value from storage where user_id IN (" + ParamStr(len(userID), 0) + ") and key = 'onesignal_player_id'"

	var params []interface{}
	for _, x := range userID {
		u := x
		params = append(params, &u)
	}

	rows, err := wx.db.Query(query, params...) //.Scan(&oneSignalPlayerId)
	if err != nil {
		return err
	}

	var playerIds []string
	for rows.Next() {
		var playerIdJson string
		err = rows.Scan(&playerIdJson)
		if err != nil {
			return err
		}
		var pid string
		err, pid = playerIdFromJSON(playerIdJson)
		if err != nil {
			return err
		}

		playerIds = append(playerIds, pid)
	}

	_, err = wx.PushNotification(playerIds, subject+`
`+content)

	return err
}

func (wx *WordX) SendPushTo(userID uuid.UUID, subject string, content string) (err error) {
	query := "SELECT value from storage where user_id = $1 and key = 'onesignal_player_id'"
	var oneSignalPlayerIdJSON *string
	err = wx.db.QueryRow(query, userID).Scan(&oneSignalPlayerIdJSON)
	if err != nil {
		return
	}

	var pid string
	err, pid = playerIdFromJSON(*oneSignalPlayerIdJSON)
	if err != nil {
		return err
	}

	_, err = wx.PushNotification([]string{pid}, subject+`
`+content)
	return
}

func (wx *WordX) PushNotification(playerIds []string, text string) (resp *http.Response, err error) {
	notificationReq := &onesignal.NotificationRequest{
		AppID:            wx.oneSignalAppID,
		Contents:         map[string]string{"en": text},
		IsIOS:            false,
		IncludePlayerIDs: playerIds,
	}
	_, res, err := wx.oneSignal.Notifications.Create(notificationReq)
	return res, err
}
