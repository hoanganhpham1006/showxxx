package adnvid

import (
	"github.com/daominah/livestream/zglobal"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"strings"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zdatabase"
)

func GetListVideoCategories() (
	map[string]interface{}, error) {
	rows, err := zdatabase.DbPool.Query(fmt.Sprintf(
		`SELECT id, name, description FROM video_categories`))
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var name, description string
		err = rows.Scan(&id, &name, &description)
		if err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"Id":          id,
			"Name":        name,
			"Description": description,
		})
	}
	return map[string]interface{}{"Rows": result}, nil
}

func GetListVideos(userId int64, limit int, offset int, orderBy string, filter string) (
	map[string]interface{}, error) {
	if misc.FindStringInSlice(orderBy, []string{
		"id", "price", "cate_id"}) == -1 {
		return nil, errors.New("Invalid orderBy")
	}
	rows, err := zdatabase.DbPool.Query(fmt.Sprintf(
		`SELECT id, name, cate_id, image, video, price, description,
    		created_time, user_id
		FROM video LEFT JOIN video_buyer ON video.id = video_buyer.video_id
		WHERE (user_id IS NULL OR user_id = $1) %v
		ORDER BY %v DESC LIMIT $2 OFFSET $3`, filter, orderBy),
		userId, limit, offset)
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	for rows.Next() {
		var id, cate_id int64
		var name, image, video, description string
		var price float64
		var created_time time.Time
		var user_id sql.NullInt64
		err = rows.Scan(&id, &name, &cate_id, &image, &video, &price,
			&description, &created_time, &user_id)
		if err != nil {
			return nil, err
		}
		hasBought := false
		if user_id.Valid {
			hasBought = true
		}
		result = append(result, map[string]interface{}{
			"Id":          id,
			"Name":        name,
			"CategoryId":  cate_id,
			"Thumbnail":   image,
			"Video":       video,
			"Price":       price,
			"Description": description,
			"CreatedTime": created_time.Format(time.RFC3339),
			"HasBought":   hasBought,
		})
	}
	return map[string]interface{}{"Rows": result}, nil
}

// dont have fields: "Video", "HasBought"
func GetListVideos2(limit int, offset int, orderBy string) (
	map[string]interface{}, error) {
	if misc.FindStringInSlice(orderBy, []string{
		"id", "price", "cate_id"}) == -1 {
		return nil, errors.New("Invalid orderBy")
	}
	rows, err := zdatabase.DbPool.Query(fmt.Sprintf(
		`SELECT id, name, cate_id, image, video, price, description, created_time
		FROM video 
		ORDER BY %v DESC LIMIT $1 OFFSET $2`, orderBy),
		limit, offset)
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	for rows.Next() {
		var id, cate_id int64
		var name, image, video, description string
		var price float64
		var created_time time.Time
		err = rows.Scan(&id, &name, &cate_id, &image, &video, &price,
			&description, &created_time)
		if err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"Id":          id,
			"Name":        name,
			"CategoryId":  cate_id,
			"Thumbnail":   image,
			"Price":       price,
			"Description": description,
			"CreatedTime": created_time.Format(time.RFC3339),
		})
	}
	return map[string]interface{}{"Rows": result}, nil
}

func GetVideoInfoById(userId int64, videoId int64) (map[string]interface{}, error) {
	row := zdatabase.DbPool.QueryRow(fmt.Sprintf(
		`SELECT v.id, v.name, v.cate_id, v.image, v.video, v.price, v.description, 
		v.is_hot, v.created_time, vb.user_id, vcb.user_id, vcb.bought_date
		FROM video v LEFT JOIN video_buyer vb ON v.id = vb.video_id
		LEFT JOIN video_categories_buyer vcb ON v.cate_id = vcb.category_id
		WHERE v.id = $1 AND ((vb.user_id IS NULL AND vcb.user_id IS NULL) OR vb.user_id = $2 OR vcb.user_id = $3)`),
		videoId, userId, userId)
	var id, cate_id int64
	var name, image, video, description, bought_category_date string
	var isHot bool
	var price float64
	var created_time time.Time
	var user_id_by_buy_video, user_id_by_buy_category sql.NullInt64
	err := row.Scan(&id, &name, &cate_id, &image, &video, &price,
		&description, &isHot, &created_time, &user_id_by_buy_video, &user_id_by_buy_category,
		&bought_category_date)
	if err != nil {
		return nil, err
	}

	hasBoughtVideo := false
	if user_id_by_buy_video.Valid {
		hasBoughtVideo = true
	}

	date_now := time.Now().Format(time.RFC3339)[0:10]

	hasBoughtCategory := false
	if (user_id_by_buy_category.Valid) && (strings.Compare(date_now, bought_category_date) == 0) {
		hasBoughtCategory = true
	}

	result := map[string]interface{}{
		"Id":          id,
		"Name":        name,
		"CategoryId":  cate_id,
		"Thumbnail":   image,
		"Video":       video,
		"Price":       price,
		"Description": description,
		"CreatedTime": created_time.Format(time.RFC3339),
		"HasBought":   hasBoughtVideo,
		"HasBoughtCategory": hasBoughtCategory,
	}
	return result, nil
}

func BuyVideo(userId int64, videoId int64) error {
	user, _ := users.GetUser(userId)
	if user == nil {
		return errors.New(l.Get(l.M022InvalidUserId))
	}
	// check has bought
	var bought_time time.Time
	row := zdatabase.DbPool.QueryRow(
		`SELECT bought_time FROM video_buyer WHERE video_id=$1 AND user_id=$2`,
		videoId, userId)
	err := row.Scan(&bought_time)
	if err == nil {
		return errors.New(l.Get(l.M044AdnvidHasBought))
	}
	// get price
	row = zdatabase.DbPool.QueryRow(`SELECT price FROM video WHERE id=$1`, videoId)
	var price float64
	err = row.Scan(&price)
	if err != nil {
		return err
	}
	// change money
	_, err = users.ChangeUserMoney(userId, users.MT_CASH, -price,
		users.REASON_BUY_VIDEO, true)
	if err != nil {
		return err
	}
	// save bought
	_, err = zdatabase.DbPool.Exec(
		"INSERT INTO video_buyer (video_id, user_id) VALUES($1, $2)", videoId, userId)
	if err != nil {
		return err
	}
	return nil
}

//AdnvidBuyCategoryDay is 
func AdnvidBuyCategoryDay(userId int64, categoryId int64) (map[string]interface {}, error) {
	now := time.Now().Format(time.RFC3339)[0:10]
	var boughtDate string
	
	row := zdatabase.DbPool.QueryRow(
		`SELECT bought_date FROM video_categories_buyer WHERE user_id=$1 AND category_id=$2`, userId, categoryId)
	err := row.Scan(&boughtDate)
	if (err == nil) && (strings.Compare(now, boughtDate) == 0) {
		return nil, errors.New("Ban da mua the loai nay trong hom nay")
	}

	_, err = users.ChangeUserMoney(userId, users.MT_CASH, -zglobal.CategoryPrice, users.REASON_BUY_VIDEO, true)
	if err != nil {
		return nil, err
	}
	
	_, err = zdatabase.DbPool.Exec(
		"INSERT INTO video_categories_buyer (category_id, user_id, bought_date) VALUES ($1, $2, $3)", categoryId, userId, now)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
