package middlewares

import "github.com/teper-ya-pomenyal/privy_stream/jwtmanager"

type MiddleWares struct {
	Auth  *authMiddleware
	Owner *ownerMiddleware
}

func NewMiddleWares(verifier *jwtmanager.Verifier) *MiddleWares {
	return &MiddleWares{
		Auth: &authMiddleware{
			verifier: verifier,
		},
		Owner: &ownerMiddleware{},
	}
}
