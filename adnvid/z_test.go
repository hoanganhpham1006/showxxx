package adnvid

import (
	"fmt"
	"testing"
	"github.com/daominah/livestream/nbackend"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	//fmt.Println("hihi")
	rows, e := GetListVideoCategories()
	// fmt.Println(rows, e)
	if e != nil || len(rows) == 0 {
		t.Error(e)
	}
}

func Test02(t *testing.T) {
	_ = fmt.Println
	//fmt.Println("hihi")
	rows, e := GetListVideos(1, 10, 0, "id",
		fmt.Sprintf(" AND cate_id = %v ", 1))
	fmt.Println(rows, e)
	if e != nil || len(rows) == 0 {
		t.Error(e)
	}
	rows, e = GetListVideos2(10, 0, "id")
	//	fmt.Println(rows, e)
	if e != nil || len(rows) == 0 {
		t.Error(e)
	}
}

func Test03(t *testing.T) {
	nbackend.InitBackend(nil) // for changing money
	err := BuyVideo(8, 2)
	_ = err
	//	fmt.Println("Test03", err)
}

func Test04(t *testing.T) {
	_, e := GetVideoInfoById(8, 2)
	if e != nil {
		t.Error(e)
	}
	_, e = GetVideoInfoById(-1, -1)
	//	fmt.Println(e)
	if e == nil {
		t.Error()
	}
}

func Test05(t *testing.T) {
	d, e := GetAdById(2)
	_ = d
	//	fmt.Println(d)
	if e != nil {
		t.Error(e)
	}
	_, e = GetVideoInfoById(-1, -1)
	//	fmt.Println(e)
	if e == nil {
		t.Error()
	}
	d, e = GetListAds(3, 0, "id")
	//	fmt.Println(d, e)
	if e != nil {
		t.Error(e)
	}
}

func Test06(t *testing.T) {
	nbackend.InitBackend(nil)
	_, e := AdnvidBuyCategoryDay(5, 1)

	if e != nil {
		t.Error(e)
	}
}
