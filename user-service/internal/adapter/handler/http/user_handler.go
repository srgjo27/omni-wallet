package http

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/omni-wallet/user-service/internal/adapter/handler/http/middleware"
	"github.com/omni-wallet/user-service/internal/adapter/handler/http/response"
	"github.com/omni-wallet/user-service/internal/core/domain"
	"github.com/omni-wallet/user-service/internal/core/services"
)

// UserHandler handles all HTTP requests related to user operations.
// It validates input and delegates business logic to the service layer.
type UserHandler struct {
	userService *services.UserService
	validate    *validator.Validate
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validate:    validator.New(),
	}
}

func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		// Public routes — no authentication required.
		users.POST("/register", h.Register)
		users.POST("/login", h.Login)

		// Protected routes — valid JWT required.
		authorized := users.Group("/")
		authorized.Use(middleware.AuthMiddleware(h.userService))
		{
			authorized.GET("/profile", h.GetProfile)
			authorized.PUT("/pin", h.SetPin)
			authorized.PUT("/kyc", h.UpdateKYC)
			authorized.POST("/logout", h.Logout)
			// Admin: list all registered users (paginated)
			authorized.GET("", h.ListUsers)
		}
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and triggers wallet creation.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      domain.RegisterRequest  true  "Registration payload"
// @Success      201   {object}  response.APIResponse
// @Failure      400   {object}  response.APIResponse
// @Failure      409   {object}  response.APIResponse
// @Router       /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	user, err := h.userService.RegisterUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			response.Conflict(c, err.Error())
			return
		}
		log.Printf("[ERROR] Register: %v", err)
		response.InternalServerError(c, "failed to register user")
		return
	}

	response.Created(c, "user registered successfully", user)
}

// Login godoc
// @Summary      Authenticate a user
// @Description  Validates credentials and returns a JWT access token.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      domain.LoginRequest  true  "Login credentials"
// @Success      200   {object}  response.APIResponse
// @Failure      400   {object}  response.APIResponse
// @Failure      401   {object}  response.APIResponse
// @Router       /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	loginResp, err := h.userService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		log.Printf("[ERROR] Login: %v", err)
		response.InternalServerError(c, "failed to process login")
		return
	}

	response.OK(c, "login successful", loginResp)
}

// ListUsers godoc
// @Summary      List all registered users (admin)
// @Description  Returns a paginated list of all users. Requires a valid JWT.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page       query  int  false  "Page number (default: 1)"
// @Param        page_size  query  int  false  "Items per page (default: 20, max: 100)"
// @Success      200  {object}  response.APIResponse
// @Failure      401  {object}  response.APIResponse
// @Router       /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.userService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		log.Printf("[ERROR] ListUsers: %v", err)
		response.InternalServerError(c, "failed to retrieve users")
		return
	}

	response.OK(c, "users retrieved successfully", gin.H{
		"users": users,
		"total": total,
		"page": page,
		"page_size": pageSize,
	})
}

// GetProfile godoc
// @Summary      Get authenticated user profile
// @Description  Returns the profile of the currently authenticated user.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.APIResponse
// @Failure      401  {object}  response.APIResponse
// @Failure      404  {object}  response.APIResponse
// @Router       /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get(middleware.AuthUserIDKey)

	user, err := h.userService.GetProfile(c.Request.Context(), userID.(string))
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		log.Printf("[ERROR] GetProfile: %v", err)
		response.InternalServerError(c, "failed to retrieve profile")
		return
	}

	response.OK(c, "profile retrieved successfully", user)
}

// SetPin godoc
// @Summary      Set or update the transaction PIN
// @Description  Sets a 6-digit numeric transaction PIN for the authenticated user.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      domain.SetPinRequest  true  "PIN payload"
// @Success      200   {object}  response.APIResponse
// @Failure      400   {object}  response.APIResponse
// @Failure      401   {object}  response.APIResponse
// @Router       /users/pin [put]
func (h *UserHandler) SetPin(c *gin.Context) {
	var req domain.SetPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	userID, _ := c.Get(middleware.AuthUserIDKey)

	if err := h.userService.SetPin(c.Request.Context(), userID.(string), req); err != nil {
		if errors.Is(err, services.ErrPinMismatch) {
			response.BadRequest(c, err.Error(), nil)
			return
		}
		log.Printf("[ERROR] SetPin: %v", err)
		response.InternalServerError(c, "failed to set transaction PIN")
		return
	}

	response.OK(c, "transaction PIN set successfully", nil)
}

// UpdateKYC godoc
// @Summary      Submit KYC information
// @Description  Submits KYC documents for the authenticated user, setting status to PENDING.
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      domain.UpdateKYCRequest  true  "KYC payload"
// @Success      200   {object}  response.APIResponse
// @Failure      400   {object}  response.APIResponse
// @Failure      401   {object}  response.APIResponse
// @Router       /users/kyc [put]
func (h *UserHandler) UpdateKYC(c *gin.Context) {
	var req domain.UpdateKYCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	userID, _ := c.Get(middleware.AuthUserIDKey)

	if err := h.userService.UpdateKYC(c.Request.Context(), userID.(string), req); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		log.Printf("[ERROR] UpdateKYC: %v", err)
		response.InternalServerError(c, "failed to update KYC information")
		return
	}

	response.OK(c, "KYC information submitted successfully", nil)
}

// Logout godoc
// @Summary      Log out the authenticated user
// @Description  Invalidates the user session in the cache.
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.APIResponse
// @Failure      401  {object}  response.APIResponse
// @Router       /users/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	userID, _ := c.Get(middleware.AuthUserIDKey)

	if err := h.userService.Logout(c.Request.Context(), userID.(string)); err != nil {
		log.Printf("[ERROR] Logout: %v", err)
		response.InternalServerError(c, "failed to logout")
		return
	}

	response.OK(c, "logout successful", nil)
}

// formatValidationErrors converts validator.ValidationErrors into a map for a clear API response.
func formatValidationErrors(err error) map[string]string {
	fieldErrors := make(map[string]string)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fieldErr := range validationErrors {
			fieldErrors[fieldErr.Field()] = buildValidationMessage(fieldErr)
		}
	}

	return fieldErrors
}

// buildValidationMessage produces a human-readable message for a single validation failure.
func buildValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters long"
	case "max":
		return "must be at most " + fe.Param() + " characters long"
	case "len":
		return "must be exactly " + fe.Param() + " characters long"
	case "numeric":
		return "must contain only numeric digits"
	default:
		return "invalid value"
	}
}
