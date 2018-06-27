package admintool

import (
	//	"fmt"

	"github.com/go-martini/martini"

	//	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zconfig"
)

func CreateRouter() *martini.ClassicMartini {
	r := martini.Classic()

	r.Get("/users/:uid", UserDetail)
	r.Put("/users/:uid", UserChangeRole) // "NewRole" string

	return r
}

func ListenAndServe() {
	r := CreateRouter()
	go r.RunOnAddr(zconfig.HttpPort)
}
