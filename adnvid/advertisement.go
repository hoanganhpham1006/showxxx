package adnvid

import (
	//	"database/sql"
	"errors"
	"fmt"
	"time"

	// l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	// "github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zdatabase"
)

func GetListAds(limit int, offset int, orderBy string) (
	map[string]interface{}, error) {
	if misc.FindStringInSlice(orderBy, []string{
		"id"}) == -1 {
		return nil, errors.New("Invalid orderBy")
	}
	rows, err := zdatabase.DbPool.Query(fmt.Sprintf(
		`SELECT id, name, url, image, "type", description, created_time
		FROM ads 
		ORDER BY %v DESC LIMIT $1 OFFSET $2`, orderBy),
		limit, offset)
	if err != nil {
		return nil, err
	}
	result := []map[string]interface{}{}
	for rows.Next() {
		var id int64
		var name, url, image, type_, description string
		var created_time time.Time
		err = rows.Scan(&id, &name, &url, &image, &type_, &description, &created_time)
		if err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"Id":          id,
			"Name":        name,
			"Url":         url,
			"Thumbnail":   image,
			"Type":        type_,
			"Description": description,
			"CreatedTime": created_time.Format(time.RFC3339),
		})
	}
	return map[string]interface{}{"Rows": result}, nil
}

func GetAdById(adId int64) (map[string]interface{}, error) {
	row := zdatabase.DbPool.QueryRow(fmt.Sprintf(
		`SELECT id, name, url, image, "type", description, created_time
		FROM ads 
		WHERE id = $1`),
		adId)
	var id int64
	var name, url, image, type_, description string
	var created_time time.Time
	err := row.Scan(&id, &name, &url, &image, &type_, &description, &created_time)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{
		"Id":          id,
		"Name":        name,
		"Url":         url,
		"Thumbnail":   image,
		"Type":        type_,
		"Description": description,
		"CreatedTime": created_time.Format(time.RFC3339),
	}
	return result, nil
}
