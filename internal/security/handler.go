package security

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gonzaloccnc/marketplace-go/pkg/httpx"
)

type HTTPAuthHandler struct {
	service   AuthService
	registrar UserRegistrar
}

func NewHTTPAuthHandler(service AuthService, registrar UserRegistrar) *HTTPAuthHandler {
	return &HTTPAuthHandler{service: service, registrar: registrar}
}

// Login authenticates a user and issues a JWT. Unknown emails and wrong
// passwords both collapse into 401 so the endpoint never reveals which emails
// are registered.
func (h *HTTPAuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteBindError(c, err)
		return
	}

	res, err := h.service.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, InvalidCredentialsError) {
			httpx.WriteError(c, http.StatusUnauthorized, InvalidCredentialsError.Error())
			return
		}
		slog.Error("login failed", "email", req.Email, "error", err)
		httpx.WriteError(c, http.StatusInternalServerError, "internal server error")
		return
	}

	httpx.WriteSuccess(c, http.StatusOK, res)
}

// Register handles public self-service user registration. It delegates the
// actual user creation to the users domain through the UserRegistrar port and
// maps the boundary errors onto HTTP status codes.
func (h *HTTPAuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteBindError(c, err)
		return
	}

	user, err := h.registrar.Register(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailAlreadyExists):
			httpx.WriteError(c, http.StatusConflict, err.Error())
		case errors.Is(err, ErrInvalidRegistration):
			httpx.WriteError(c, http.StatusBadRequest, err.Error())
		default:
			slog.Error("registration failed", "email", req.Email, "error", err)
			httpx.WriteError(c, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httpx.WriteSuccess(c, http.StatusCreated, user)
}
