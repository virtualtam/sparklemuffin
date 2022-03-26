package www

import (
	"net/http"

	"github.com/virtualtam/yawbe/pkg/user"
)

func login(userService *user.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {}
}
