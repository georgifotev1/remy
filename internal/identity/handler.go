package identity

import "net/http"

type IdentityHandlers struct{}

func NewIdentityHandlers() *IdentityHandlers {
	return &IdentityHandlers{}
}

func (h *IdentityHandlers) SignUp(w http.ResponseWriter, r *http.Request) {}
