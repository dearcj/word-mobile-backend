package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/heroiclabs/nakama/api"
	"github.com/heroiclabs/nakama/shortid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"math"
	"strconv"
	"strings"
	"time"
)

type Words struct {
	Words  []string `json:"words"`
	Points []string `json:"points"`
	Awards []int    `json:"awards"`
}

type Category struct {
	Locked       bool      `protobuf:"varint,6,opt,name=locked,proto3" json:"locked"`
	Id           uuid.UUID `protobuf:"bytes,3,opt,name=id,proto3" json:"id"`
	Name         string    `protobuf:"bytes,4,opt,name=name,proto3" json:"name"`
	Languages    []*Lang   `protobuf:"bytes,1,rep,name=languages,proto3" json:"languages,omitempty"`
	UnlockPrice  int32     `protobuf:"varint,2,opt,name=unlock_price,json=unlockPrice,proto3" json:"unlock_price"`
	ImageData    string    `protobuf:"bytes,5,opt,name=imageData,proto3" json:"imageData"`
	ImageDataURL string    `protobuf:"bytes,5,opt,name=imageData,proto3" json:"imageDataURL"`
	Description  string    `json:"description"`
	Words        *Words    `json:"words,omitempty"`
}

func AwardsFromWords(words []string) []int {
	var res []int
	for _, x := range words {
		l := len(x)
		v := (l / 2)
		if v == 0 {
			v = 1
		}

		if v > 8 {
			v = 8
		}

		res = append(res, v)
	}

	return res
}

func (x *Category) SetLanguageWords(lang int) {
	x.Words = &Words{}
	x.Words.Words = x.Languages[lang].Words
	x.Words.Points = x.Languages[lang].Points
	x.Words.Awards = AwardsFromWords(x.Words.Words)
}

type SwagLevel struct {
	Name   string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Points int32  `protobuf:"varint,2,opt,name=points,proto3" json:"points,omitempty"`
	Desc   string `json:"desc,omitempty"`
}

type Settings struct {
	CoinsPerPoint       float32         `protobuf:"fixed32,3,opt,name=coinsPerPoint,proto3" json:"coinsPerPoint"`
	SwagLevels          []*SwagLevel    `protobuf:"bytes,4,rep,name=swagLevels,proto3" json:"swagLevels"`
	Categories          *CategoriesList `protobuf:"bytes,1,opt,name=categories,proto3" json:"categories"`
	AvailableCategories []string        `protobuf:"bytes,2,rep,name=availableCategories,proto3" json:"availableCategories"`
	H2HWords            int32           `json:"h2hwords"`
	Langs               []string        `json:"langs"`
	WordShowupTime      float32         `json:"word_showup_time"`
	LenRound            int32           `json:"len_round"`
	TimePenaltySec      int32           `json:"time_penalty_sec"`
}

type CategoriesList struct {
	Categories []*Category `protobuf:"bytes,1,rep,name=categories,proto3" json:"categories"`
}

type Lang struct {
	Name   string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
	Words  []string `protobuf:"bytes,2,rep,name=words,proto3" json:"words"`
	Points []string `protobuf:"bytes,3,rep,name=points,proto3" json:"points"`
	Awards []int    `json:"awards"`
}

var WX_LANGS = []string{"english", "spanish", "italian", "french", "portuguese"}

func DefLangs(c *Category) {
	c.Languages = DefLangsList()
}

func DefLangsList() []*Lang {
	var l []*Lang
	for _, x := range WX_LANGS {
		l = append(l, &Lang{
			Name: x,
		})
	}

	return l
}

func ScanCat(r *sql.Rows, language *string, scanImages bool) (*Category, error) {
	cat := &Category{}
	DefLangs(cat)

	if language != nil {
		var outerinx int = -1
		for inx, x := range cat.Languages {
			if x.Name == *language {
				outerinx = inx
			}
		}

		if outerinx == -1 {
			return nil, StatusError(codes.NotFound, "No such language", nil)
		}

		err := r.Scan(&cat.Name,
			pq.Array(&cat.Languages[outerinx].Words),
			&cat.UnlockPrice)

		if err != nil {
			return nil, err
		}

	} else {
		var str sql.NullString
		args := []interface{}{
			&cat.Id,
			&cat.Name,
			pq.Array(&cat.Languages[0].Words),
			pq.Array(&cat.Languages[1].Words),
			pq.Array(&cat.Languages[2].Words),
			pq.Array(&cat.Languages[3].Words),
			pq.Array(&cat.Languages[4].Words),
			pq.Array(&cat.Languages[0].Points),
			pq.Array(&cat.Languages[1].Points),
			pq.Array(&cat.Languages[2].Points),
			pq.Array(&cat.Languages[3].Points),
			pq.Array(&cat.Languages[4].Points),
			&cat.Description,
			&cat.UnlockPrice,
		}

		if scanImages == true {
			args = append(args, &str)
		}

		err := r.Scan(args...)
		cat.ImageDataURL = "/image/" + cat.Id.String()

		for _, l := range cat.Languages {
			l.Awards = AwardsFromWords(l.Words)
		}

		if str.Valid {
			cat.ImageData = str.String
		}

		if err != nil {
			return nil, err
		}
	}

	return cat, nil

}

func CalculateWordPoints(s string) (score int) {
	lower := strings.ToLower(s)
	for _, x := range lower {
		switch x {
		case 'a', 'e', 'i', 'o', 'u':
			score += 2
			break
		case 'w', 'y':
			score += 1
			break
		default:
			score -= 1
		}

	}
	return
}

func (s *ApiServer) IsAdmin(userID uuid.UUID) bool {
	query := "SELECT role from users where id = $1;"
	var role uint32
	err := s.db.QueryRow(query, userID.String()).Scan(&role)
	if err != nil {

	}
	return role == 1
}

func stripWords(cat *Category) {
	for _, x := range cat.Languages {
		x.Words = []string{}
		x.Points = []string{}
	}
}

func (s *ApiServer) GetSingleCategory(id uuid.UUID, onlyUnlockedForUserID *uuid.UUID, getImage bool) (*Category, error) {
	locked := false
	if onlyUnlockedForUserID != nil {
		query := "SELECT category_id from unlocked_categories where user_id = $1 AND category_id = $2;"

		err := s.db.QueryRow(query, onlyUnlockedForUserID, &id).Scan()
		if err != nil {
			locked = true
		}
	}

	var getImageStr string
	if getImage {
		getImageStr = ", image_data "
	}
	query := "SELECT id, catname, words_eng, words_spa, words_ita, words_fre, words_por, points_eng, points_spa, points_ita, points_fre, points_por, description, unlock_price " + getImageStr + " from category where id = $1;"
	rows, err := s.db.Query(query, &id)

	if err != nil {
		return nil, err
	}

	rows.Next()
	cat, err := ScanCat(rows, nil, getImage)

	cat.Locked = locked
	if cat.Locked {
		stripWords(cat)
	}

	if err != nil {
		return nil, err
	}

	return cat, nil
}

type DashboardUserStorage struct {
	Collection string
	Key        string
	Read       int32
	Write      int32
	Value      string
}

type DashboardUser struct {
	Id            uuid.UUID
	Username      string
	Email         *string
	Facebook_Id   *string
	Google_Id     *string
	Gamecenter_Id *string
	Steam_Id      *string
	Custom_Id     *string
	Create_Time   time.Time
	Role          int32
	StorageData   []*DashboardUserStorage
}

type PurchaseItem struct {
	Receipt string `json:"receipt"`
	Amount  int    `json:"amount"`
	Item    string `json:"item"`
}

func (s *ApiServer) TrackPurchase(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Items []PurchaseItem `json:"items"`
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if len(b.Payload.Items) == 0 {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	var vals []string
	var params []interface{}
	for inx, x := range b.Payload.Items {
		v := "(" + ParamStr(4, inx*4) + ")"
		vals = append(vals, v)
		params = append(params, &x.Item, &x.Amount, &x.Receipt, &userID)
	}
	query := "insert into purchases (item, amount, receipt, user_id) values " + strings.Join(vals, ",") + ";"

	_, err = s.db.Exec(query,
		params...)

	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "Purchase added",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) CountryList(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	query := `select country from countries;`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	var countries []string
	for rows.Next() {
		var countryName string
		err := rows.Scan(&countryName)
		if err != nil {
			return nil, err
		}
		countries = append(countries, countryName)
	}

	bytea, err := json.Marshal(&countries)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) GetRestoreCode(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Email string `json:"email"`
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	query := "SELECT id from users where email = $1;"
	var userId uuid.UUID

	err = s.db.QueryRow(query, &b.Email).Scan(&userId)
	if err != nil {
		return nil, StatusError(codes.Internal, "No user with this email", nil)
	}

	sid, _ := shortid.New(1, shortid.DefaultABC, 22)
	code, _ := sid.Generate()

	query = "INSERT INTO passreq (user_id, code) VALUES ($1, $2);"

	s.runtimePool.wx.SendPasswordEmail(b.Email, code)

	_, err = s.db.Exec(query, &userId, &code)
	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "Code sent by email",
	}

	bytea, err := json.Marshal(&res)
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (w *WordX) CheckCode(c string) (bool, *uuid.UUID, *uuid.UUID) {
	query := "SELECT code, id, user_id FROM passreq WHERE code = $1 "
	var code *string
	var id *uuid.UUID
	var userid *uuid.UUID
	err := w.apiServ.db.QueryRow(query, &c).Scan(&code, &id, &userid)
	res := !(err != nil || code == nil)
	return res, id, userid
}

func (s *ApiServer) CheckCode(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Code string `json:"code"`
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	havecode, _, _ := s.runtimePool.wx.CheckCode(b.Code)
	if !havecode {
		return nil, StatusError(codes.NotFound, "No such request code", nil)
	} else {
		res := struct {
			Message string
		}{
			Message: "Code is correct",
		}

		bytea, err := json.Marshal(&res)
		return &api.WordXAPIRes{Payload: string(bytea)}, err
	}
}

func (s *ApiServer) RestorePassword(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		NewPassword string `json:"new_password"`
		Code        string `json:"code"`
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if len(b.NewPassword) < 8 {
		return nil, StatusError(codes.InvalidArgument, "Password must be longer than 8 characters.", nil)
	}

	havecode, id, user_id := s.runtimePool.wx.CheckCode(b.Code)
	if havecode == false {
		return nil, StatusError(codes.Internal, "Wrong restoration code", nil)
	} else {
		query := "DELETE from passreq where id = $1;"

		_, err := s.db.Exec(query, &id)
		if err != nil {
			return nil, err
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(b.NewPassword), bcrypt.DefaultCost)
		query = `UPDATE users SET (password) = ($1) where id = $2;`
		_, err = s.db.Exec(query, &hashedPassword, &user_id)
		if err != nil {
			return nil, err
		}

		res := struct {
			Message string
		}{
			Message: "User updated",
		}

		bytea, err := json.Marshal(&res)
		if err != nil {
			return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
		} else {
			return &api.WordXAPIRes{Payload: string(bytea)}, nil
		}
	}
}

func (s *ApiServer) DashboardUpdateUser(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			User *DashboardUser
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.User == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	query := `UPDATE users SET (email, username, facebook_id, google_id, gamecenter_id, steam_id, custom_id, role) = ($2, $3, $4, $5, $6, $7, $8, $9) where id = $1;`

	_, err = s.db.Exec(query,
		&b.Payload.User.Id,
		&b.Payload.User.Email,
		&b.Payload.User.Username,
		&b.Payload.User.Facebook_Id,
		&b.Payload.User.Google_Id,
		&b.Payload.User.Gamecenter_Id,
		&b.Payload.User.Steam_Id,
		&b.Payload.User.Custom_Id,
		&b.Payload.User.Role)

	if err != nil {
		return nil, err
	}
	writeObj := make(map[uuid.UUID][]*api.WriteStorageObject)

	for _, x := range b.Payload.User.StorageData {
		writeObj[b.Payload.User.Id] = append(writeObj[b.Payload.User.Id], &api.WriteStorageObject{
			Collection:      x.Collection,
			Key:             x.Key,
			Value:           StorageValueSet(x.Value),
			PermissionRead:  &wrappers.Int32Value{Value: x.Read},
			PermissionWrite: &wrappers.Int32Value{Value: x.Write},
		})

	}

	_, _, err = StorageWriteObjects(s.logger, s.db, true, writeObj)

	if err != nil {
		return nil, err
	}

	res := struct {
		Message string
	}{
		Message: "User updated",
	}

	bytea, err := json.Marshal(&res)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) DashboardGetUserWithStorage(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			User_id *string
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.User_id == nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	}

	query := "SELECT username, email, facebook_id, google_id, gamecenter_id, steam_id, custom_id, create_time, role FROM users WHERE id = $1 "
	row := s.db.QueryRow(query, &b.Payload.User_id)
	u := &DashboardUser{}
	u.Id, err = uuid.FromString(*b.Payload.User_id)

	if err != nil {
		return nil, err
	}

	err = row.Scan(
		&u.Username,
		&u.Email,
		&u.Facebook_Id,
		&u.Google_Id,
		&u.Gamecenter_Id,
		&u.Steam_Id,
		&u.Custom_Id,
		&u.Create_Time,
		&u.Role,
	)

	if err != nil {
		return nil, err
	}

	query = "SELECT collection, key, value, read, write FROM storage WHERE user_id = $1 "
	rows, err := s.db.Query(query, &b.Payload.User_id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		st := &DashboardUserStorage{}
		err := rows.Scan(&st.Collection, &st.Key, &st.Value, &st.Read, &st.Write)

		if err != nil {
			return nil, err
		}

		err, st.Value = StorageValueGet(st.Value)
		if err != nil {
			return nil, err
		}

		u.StorageData = append(u.StorageData, st)
	}

	bytea, err := json.Marshal(&u)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) GetAchievements(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	query := "SELECT id, name, event, condition, reward FROM achs"
	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	type Ach struct {
		Id        string `json:"id"`
		Name      string `json:"name"`
		Event     string `json:"event"`
		Condition int    `json:"condition"`
		Reward    int    `json:"reward"`
	}

	var arr []*Ach
	for rows.Next() {
		var a Ach
		err := rows.Scan(&a.Id, &a.Name, &a.Event, &a.Condition, &a.Reward)

		if err != nil {
			return nil, err
		}

		arr = append(arr, &a)
	}

	var res struct {
		Achievements []*Ach
	}

	res.Achievements = arr

	bytea, err := json.Marshal(res)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) DashboardGetUsers(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			PageSize     int
			Filter_email string
			Filter_id    string
			Page         int
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	query := "SELECT id, username, email, facebook_id, google_id, gamecenter_id, steam_id, custom_id, create_time, role FROM users"
	var users []*DashboardUser
	var args []interface{}

	var whereCond []string
	whereCond = append(whereCond, ` id != `+` '00000000-0000-0000-0000-000000000000' `)

	if b.Payload.Filter_id != "" {
		whereCond = append(whereCond, " id = $1 ")
		args = append(args, &b.Payload.Filter_id)
	}

	if b.Payload.Filter_email != "" {
		if len(whereCond) > 1 {
			whereCond = append(whereCond, " email = $2 ")
		} else {
			whereCond = append(whereCond, " email = $1 ")
		}
		args = append(args, &b.Payload.Filter_email)
	}

	var fullWhere string
	if len(whereCond) > 0 {
		fullWhere = " WHERE " + strings.Join(whereCond, " AND ")
	}

	countQuery := "SELECT count(1) FROM users " + fullWhere
	var count int
	err = s.db.QueryRow(countQuery, args...).Scan(&count)

	if err != nil {
		return nil, err
	}

	repl := strings.NewReplacer("$1", "id",
		"$2", strconv.Itoa(b.Payload.PageSize),
		"$3", strconv.Itoa((b.Payload.Page)*b.Payload.PageSize))

	pagination := repl.Replace(" ORDER BY $1 LIMIT $2 OFFSET $3")
	totalQ := query + fullWhere + pagination
	rows, err := s.db.Query(totalQ, args...)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		u := &DashboardUser{}

		err := rows.Scan(&u.Id,
			&u.Username,
			&u.Email,
			&u.Facebook_Id,
			&u.Google_Id,
			&u.Gamecenter_Id,
			&u.Steam_Id,
			&u.Custom_Id,
			&u.Create_Time,
			&u.Role,
		)

		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err != nil {
		return nil, err
	}

	var res struct {
		Users []*DashboardUser
		Pages int
	}

	res.Users = users
	res.Pages = int(math.Ceil(float64(count) / float64(b.Payload.PageSize)))
	bytea, err := json.Marshal(res)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) DashboardUpdateCategory(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Category *Category
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.Category == nil || len(b.Payload.Category.Languages) < 4 {
		return nil, StatusError(codes.Internal, "Wrong parameters", nil)
	}

	for _, x := range b.Payload.Category.Languages {
		points := []string{}
		for _, w := range x.Words {
			points = append(points, strconv.Itoa(CalculateWordPoints(w)))
		}

		x.Points = points
		x.Awards = AwardsFromWords(x.Words)
	}

	query := "UPDATE category SET (catname, words_eng, words_spa, words_ita, words_fre, words_por, points_eng, points_spa, points_ita, points_fre, points_por, description, unlock_price, image_data) = ($2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) where id = $1;"

	_, err = s.db.Exec(query,
		&b.Payload.Category.Id,
		&b.Payload.Category.Name,
		pq.Array(&b.Payload.Category.Languages[0].Words),
		pq.Array(&b.Payload.Category.Languages[1].Words),
		pq.Array(&b.Payload.Category.Languages[2].Words),
		pq.Array(&b.Payload.Category.Languages[3].Words),
		pq.Array(&b.Payload.Category.Languages[4].Words),
		pq.Array(&b.Payload.Category.Languages[0].Points),
		pq.Array(&b.Payload.Category.Languages[1].Points),
		pq.Array(&b.Payload.Category.Languages[2].Points),
		pq.Array(&b.Payload.Category.Languages[3].Points),
		pq.Array(&b.Payload.Category.Languages[4].Points),
		&b.Payload.Category.Description,
		&b.Payload.Category.UnlockPrice,
		&b.Payload.Category.ImageData)

	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(b.Payload.Category)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) UnlockCategory(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Category_id *uuid.UUID
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.Category_id == nil {
		return nil, StatusError(codes.NotFound, "JSON Unmarshal error", nil)
	}

	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)
	cat, err := s.GetSingleCategory(*b.Payload.Category_id, nil, false)

	if err != nil {
		return nil, err
	}

	objectIDs := []*api.ReadStorageObjectId{
		{
			Key:        "money",
			Collection: "resources",
			UserId:     userID.String(),
		},
	}
	res, err := StorageReadObjects(s.logger, s.db, userID, objectIDs)
	if err != nil {
		return nil, err
	}

	if len(res.Objects) == 0 {
		return nil, StatusError(codes.NotFound, "No resource storage for this user", nil)
	}

	resources := res.Objects[0]
	_, v := StorageValueGet(resources.Value)
	money, err := strconv.ParseInt(v, 10, 0)
	if err != nil {
		return nil, err
	}

	if int32(money) >= cat.UnlockPrice {
		resources.Value = strconv.Itoa(int(money) - int(cat.UnlockPrice))
		writeObj := make(map[uuid.UUID][]*api.WriteStorageObject)
		writeObj[userID] = []*api.WriteStorageObject{{
			Version:         resources.Version,
			Collection:      resources.Collection,
			Key:             resources.Key,
			Value:           StorageValueSet(resources.Value),
			PermissionRead:  &wrappers.Int32Value{Value: 1},
			PermissionWrite: &wrappers.Int32Value{Value: 0},
		}}

		_, code, err := StorageWriteObjects(s.logger, s.db, true, writeObj)
		if err != nil {
			return nil, StatusError(code, "Transaction failed", err)
		}
	} else {
		return nil, StatusError(codes.NotFound, "Not enough money", nil)
	}

	query := "INSERT INTO unlocked_categories (user_id, category_id) VALUES ($1, $2);"

	_, err = s.db.Exec(query, &userID, &b.Payload.Category_id)
	if err != nil {
		return nil, err
	}

	cat.SetLanguageWords(0)
	//cat.Languages = nil

	bytea, err := json.Marshal(cat)
	if err != nil {
		return nil, StatusError(codes.Internal, "JSON Marhsal error", nil)
	} else {
		return &api.WordXAPIRes{Payload: string(bytea)}, nil
	}
}

func (s *ApiServer) GetSettingsInner() (*Settings, error) {
	query := "SELECT  swag_levels, swag_levels_points,swag_levels_desc, coins_per_point, h2hwords, word_showup_time, len_round,time_penalty_sec from settings where id = '00000000-0000-0000-0000-000000000000';"
	set := &Settings{}
	floatArray := []int64{}
	levNames := []string{}
	descArray := []string{}
	err := s.db.QueryRow(query).Scan(pq.Array(&levNames), pq.Array(&floatArray), pq.Array(&descArray), &set.CoinsPerPoint, &set.H2HWords, &set.WordShowupTime, &set.LenRound, &set.TimePenaltySec)
	if err != nil {
		return nil, err
	}

	for inx, x := range levNames {
		desc := ""
		if len(descArray) > inx {
			desc = descArray[inx]
		}

		set.SwagLevels = append(set.SwagLevels, &SwagLevel{
			Points: int32(floatArray[inx]),
			Name:   x,
			Desc:   desc,
		})
	}

	set.Langs = WX_LANGS

	if err != nil {
		return nil, err
	} else {
		return set, nil
	}

}

func (s *ApiServer) GetCategory(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Category_id *uuid.UUID
		}
	}

	if b.Payload.Category_id == nil {
		return nil, StatusError(codes.NotFound, "JSON Unmarshal error", nil)
	}

	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	cat, err := s.GetSingleCategory(*b.Payload.Category_id, &userID, true)

	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(&cat)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardGetSettings(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	set, err := s.GetSettingsInner()

	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(&set)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardUpdateSettings(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload *Settings
	}

	err := json.Unmarshal([]byte(in.Payload), &b)

	if b.Payload == nil {
		return nil, StatusError(codes.NotFound, "JSON Unmarshal error", nil)
	}

	if err != nil {
		return nil, err
	}

	query := "UPDATE settings SET (swag_levels, swag_levels_points, swag_levels_desc, coins_per_point, h2hwords, word_showup_time, len_round, time_penalty_sec) = ($1, $2, $3, $4, $5, $6, $7, $8) where id = '00000000-0000-0000-0000-000000000000';"
	var floatArray []int64
	var descArray []string
	var levNames []string
	for _, x := range b.Payload.SwagLevels {
		descArray = append(descArray, x.Desc)
		floatArray = append(floatArray, int64(x.Points))
		levNames = append(levNames, x.Name)
	}

	_, err = s.db.Exec(query, pq.Array(&levNames), pq.Array(&floatArray), pq.Array(&descArray), &b.Payload.CoinsPerPoint, &b.Payload.H2HWords, &b.Payload.WordShowupTime, &b.Payload.LenRound, &b.Payload.TimePenaltySec)
	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(struct{ msg string }{msg: "Settings update"})
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardGetCategory(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Category_id *uuid.UUID
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	if b.Payload.Category_id == nil {
		return nil, StatusError(codes.NotFound, "JSON Unmarshal error", nil)
	}

	cat, err := s.GetSingleCategory(*b.Payload.Category_id, nil, true)

	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(&cat)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardGetCategories(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	query := "SELECT id, catname, words_eng, words_spa, words_ita, words_fre, words_por, points_eng, points_spa, points_ita, points_fre, points_por, description, unlock_price from category;"
	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	var results = CategoriesList{}

	for rows.Next() {
		cat, err := ScanCat(rows, nil, false)

		if err != nil {
			return nil, err
		}

		results.Categories = append(results.Categories, cat)
	}

	bytea, err := json.Marshal(&results)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardAddCategory(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	uid, err := uuid.NewV4()
	//userID := uuid.Must(uid, err).String()

	query := "INSERT INTO category (id, catname, words_eng, words_spa, words_ita, words_fre, words_por, description, unlock_price) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);"
	catname := ""
	words := "{}"
	unlock_price := 0
	desc := ""
	_, err = s.db.Exec(query, &uid, &catname, &words, &words, &words, &words, &words, &desc, &unlock_price)

	if err != nil {
		return nil, err
	}

	results := &Category{
		Id: uid,
	}
	DefLangs(results)

	bytea, err := json.Marshal(&results)
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardDeleteCategory(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			Category_id uuid.UUID
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	query := "DELETE from category where id = $1;"
	_, err = s.db.Exec(query, b.Payload.Category_id)

	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(struct{ msg string }{msg: "Category deleted"})
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) AddResource(userid uuid.UUID, points int32, coins int32) error {
	objectIDs := []*api.ReadStorageObjectId{
		{
			Key:        "money",
			Collection: "resources",
			UserId:     userid.String(),
		},
		{
			Key:        "points",
			Collection: "resources",
			UserId:     userid.String(),
		},
	}

	if coins < 0 {
		coins = 0
	}

	res, err := StorageReadObjects(s.logger, s.db, userid, objectIDs)
	if err != nil {
		return err
	}

	var objMoney, objPoints *api.StorageObject

	for _, x := range res.Objects {
		if x.Key == "money" {
			objMoney = x
		}

		if x.Key == "points" {
			objPoints = x
		}
	}

	writeObj := make(map[uuid.UUID][]*api.WriteStorageObject)

	if coins != 0 {
		if objMoney != nil {
			_, v := StorageValueGet(objMoney.Value)
			prevmoney, err := strconv.ParseInt(v, 10, 0)

			if err != nil {
				return err
			}

			objMoney.Value = strconv.Itoa(int(prevmoney) + int(coins))
			writeObj[userid] = append(writeObj[userid], &api.WriteStorageObject{
				Version:         objMoney.Version,
				Collection:      objMoney.Collection,
				Key:             objMoney.Key,
				Value:           StorageValueSet(objMoney.Value),
				PermissionRead:  &wrappers.Int32Value{Value: 1},
				PermissionWrite: &wrappers.Int32Value{Value: 0},
			})
		} else {
			writeObj[userid] = append(writeObj[userid], &api.WriteStorageObject{
				Collection:      "resources",
				Key:             "money",
				Value:           StorageValueSet(strconv.Itoa(int(coins))),
				PermissionRead:  &wrappers.Int32Value{Value: 1},
				PermissionWrite: &wrappers.Int32Value{Value: 0},
			})
		}
	}

	if points != 0 {
		if objPoints != nil {
			_, v := StorageValueGet(objPoints.Value)
			prevpoints, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				return err
			}

			newpoints := strconv.Itoa(int(prevpoints) + int(points))
			writeObj[userid] = append(writeObj[userid], &api.WriteStorageObject{
				Version:         objPoints.Version,
				Collection:      objPoints.Collection,
				Key:             objPoints.Key,
				Value:           StorageValueSet(newpoints),
				PermissionRead:  &wrappers.Int32Value{Value: 1},
				PermissionWrite: &wrappers.Int32Value{Value: 0},
			})
		} else {
			writeObj[userid] = append(writeObj[userid], &api.WriteStorageObject{
				Collection:      "resources",
				Key:             "points",
				Value:           StorageValueSet(strconv.Itoa(int(points))),
				PermissionRead:  &wrappers.Int32Value{Value: 1},
				PermissionWrite: &wrappers.Int32Value{Value: 0},
			})
		}
	}

	_, code, err := StorageWriteObjects(s.logger, s.db, true, writeObj)

	if err != nil {
		return StatusError(code, "Transaction failed", err)
	}

	return nil
}

func (s *ApiServer) ResolveRound(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			AddedPoints int32
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	var res struct {
		AddedPoints int32
		AddedCoins  int32
	}
	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	res.AddedCoins = int32(float32(b.Payload.AddedPoints) * s.runtimePool.wx.Cache.Settings.CoinsPerPoint)
	res.AddedPoints = b.Payload.AddedPoints

	err = s.AddResource(userID, res.AddedPoints, res.AddedCoins)

	if err != nil {
		return nil, err
	}

	bytea, _ := json.Marshal(&res)
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) DashboardDeleteUser(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	var b struct {
		Payload struct {
			User_id uuid.UUID
		}
	}

	err := json.Unmarshal([]byte(in.Payload), &b)
	if err != nil {
		return nil, err
	}

	query := "DELETE from users where id = $1;"
	_, err = s.db.Exec(query, &b.Payload.User_id)

	if err != nil {
		return nil, err
	}

	bytea, err := json.Marshal(struct{ msg string }{msg: "User deleted"})
	if err != nil {
		return nil, err
	}
	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) GetAvailCategories(userID uuid.UUID) ([]string, error) {
	query := "SELECT category_id from unlocked_categories where user_id = $1;"

	var result []string
	rows, err := s.db.Query(query, &userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var str string
		if err := rows.Scan(&str); err != nil {
			return nil, err
		}
		result = append(result, str)
	}

	return result, nil
}

func (s *ApiServer) GetSettings(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	results, err := s.GetSettingsInner()

	if err != nil {
		return nil, err
	}

	userID := ctx.Value(ctxUserIDKey{}).(uuid.UUID)

	query := "SELECT id, catname, words_eng, words_spa, words_ita, words_fre, words_por,  points_eng, points_spa, points_ita, points_fre, points_por, description, unlock_price from category;"
	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	results.Categories = &CategoriesList{}
	avail, err := s.GetAvailCategories(userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		cat, err := ScanCat(rows, nil, false)
		if err != nil {
			return nil, err
		}

		if cat.UnlockPrice != 0 {
			found := false
			for _, x := range avail {
				if x == cat.Id.String() {
					found = true
				}
			}

			if !found {
				cat.Locked = true
				stripWords(cat)
			}
		}

		results.Categories.Categories = append(results.Categories.Categories, cat)
	}

	for _, x := range results.Categories.Categories {
		x.SetLanguageWords(0)
		//	x.Languages = nil
	}

	results.AvailableCategories = avail
	bytea, err := json.Marshal(&results)
	if err != nil {
		return nil, err
	}

	return &api.WordXAPIRes{Payload: string(bytea)}, nil
}

func (s *ApiServer) GetCategories(ctx context.Context, in *api.PayloadParams) (*api.WordXAPIRes, error) {
	return s.DashboardGetCategories(ctx, in)
}

func InitWordXAPI(wx *WordX) {

	wx.AddCallbackRPC("clientrpc.getsettings", func(queryParams map[string][]string, uid string, username string, sessionExpiry int64, sid string, payload interface{}) (result interface{}, err error, code codes.Code) {
		bytearr, err := json.Marshal(wx.Cache.Settings)

		if err != nil {
			return nil, err, codes.Internal
		}

		return bytearr, nil, codes.OK
	})

}
