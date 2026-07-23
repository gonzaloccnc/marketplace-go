package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gonzaloccnc/marketplace-go/internal/token"
	"github.com/gonzaloccnc/marketplace-go/pkg/httpx"
	"github.com/google/uuid"
)

type HTTPUserHandler struct {
	service UserService
}

func NewHTTPUserHandler(service UserService) *HTTPUserHandler {
	return &HTTPUserHandler{service: service}
}

// authActorID extracts the authenticated actor's id from the JWT claims injected
// by the security middleware, writing a 401 and returning ok=false when the
// request is unauthenticated or carries an unparseable subject.
func authActorID(c *gin.Context) (uuid.UUID, bool) {
	actor, ok := token.ClaimsFromContext(c)
	if !ok {
		httpx.WriteError(c, http.StatusUnauthorized, "unauthenticated")
		return uuid.Nil, false
	}

	id, err := uuid.Parse(actor.Subject)
	if err != nil {
		httpx.WriteError(c, http.StatusUnauthorized, "invalid actor")
		return uuid.Nil, false
	}

	return id, true
}

// pathUserID parses the :id path parameter as a UUID, writing a 400 and
// returning ok=false when it is malformed.
func pathUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.WriteError(c, http.StatusBadRequest, "invalid user id")
		return uuid.Nil, false
	}
	return id, true
}

// CreateUser creates a user on behalf of an authenticated caller. The acting
// user is taken from the JWT claims injected by the security middleware, so we
// audit who created the record (created_via = admin, created_by = actor).
func (h *HTTPUserHandler) CreateUser(c *gin.Context) {
	actorID, ok := authActorID(c)
	if !ok {
		return
	}

	var req AdminUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteBindError(c, err)
		return
	}

	created, err := h.service.CreateUserByAdmin(c, &req, actorID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	// new user; swap this for an email / password-reset flow in production.
	httpx.WriteSuccess(c, http.StatusCreated, created.User)
}

// GetUserByID returns a single user by id.
func (h *HTTPUserHandler) GetUserByID(c *gin.Context) {
	userID, ok := pathUserID(c)
	if !ok {
		return
	}

	user, err := h.service.GetUserByID(c, userID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	httpx.WriteSuccess(c, http.StatusOK, user)
}

// UpdateUser updates an existing user. The acting user is taken from the JWT
// claims injected by the security middleware and recorded as updated_by for
// audit.
func (h *HTTPUserHandler) UpdateUser(c *gin.Context) {
	actorID, ok := authActorID(c)
	if !ok {
		return
	}

	userID, ok := pathUserID(c)
	if !ok {
		return
	}

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.WriteBindError(c, err)
		return
	}

	updated, err := h.service.UpdateUser(c, userID, &req, actorID)
	if err != nil {
		h.writeServiceError(c, err)
		return
	}

	httpx.WriteSuccess(c, http.StatusOK, updated)
}

// DeleteUser soft-deletes a user by id. The acting user is taken from the JWT
// claims injected by the security middleware and recorded as deleted_by for
// audit.
func (h *HTTPUserHandler) DeleteUser(c *gin.Context) {
	actorID, ok := authActorID(c)
	if !ok {
		return
	}

	userID, ok := pathUserID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteUser(c, userID, actorID); err != nil {
		h.writeServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// writeServiceError maps a users service error onto the standard error envelope
// with the appropriate HTTP status, so every handler reports failures the same
// way.
func (h *HTTPUserHandler) writeServiceError(c *gin.Context, err error) {
	switch err {
	case UserNotFoundError:
		httpx.WriteError(c, http.StatusNotFound, err.Error())
	case EmailAlreadyExistsError:
		httpx.WriteError(c, http.StatusConflict, err.Error())
	case NameTooShortError, PasswordInsufficientError:
		httpx.WriteError(c, http.StatusBadRequest, err.Error())
	default:
		httpx.WriteError(c, http.StatusInternalServerError, err.Error())
	}
}
