package admintool

import (
	//	"fmt"

	"github.com/go-martini/martini"

	//	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zconfig"
)

func CreateRouter() *martini.ClassicMartini {
	r := martini.Classic()

	r.Post("/users/login", UserLogin) // "Username" string, "Password" string

	r.Get("/users/:uid", UserDetail)            //
	r.Put("/users/:uid/role", UserChangeRole)   // "NewRole" string
	r.Put("/users/:uid/suspend", UserSuspend)   // "IsSuspended" bool
	r.Patch("/users/:uid/cash", UserChangeCash) // "Change" float64

	r.Get("/stat/online", UserOnlineStat)

	return r
}

func ListenAndServe() {
	r := CreateRouter()
	go r.RunOnAddr(zconfig.HttpPort)
}
