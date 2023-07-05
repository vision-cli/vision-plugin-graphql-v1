package auth

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	authorization = "Authorization"
	bearer        = "Bearer "
)

type AuthenticationService interface {
	Authenticate(handler http.Handler) (http.Handler, error)
}

type Service struct {
	client *http.Client
}

func NewService() *Service {
	return &Service{
		client: &http.Client{},
	}
}

type AuthUser struct {
	Email string
}

// Authenticate parses the JWT token if it has been supplied
// and if valid, puts it in the request context.
func (s *Service) Authenticate(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get token from request
		reqToken := strings.TrimPrefix(r.Header.Get(authorization), bearer)
		if reqToken == "" {
			http.Error(w, "failed to get jwt token", http.StatusInternalServerError)
			return
		}

		// TODO parse token and check signing method
		_, err := s.authenticate(reqToken)
		if err != nil {
			http.Error(w, "failed to authenticate token", http.StatusInternalServerError)
			return
		}

		// TODO map claims of token

		handler.ServeHTTP(w, r)
	})
}

func (s *Service) authenticate(reqToken string) (*jwt.Token, error) {
	// TODO parse token, checking signing method and keys
	return &jwt.Token{}, nil
}
