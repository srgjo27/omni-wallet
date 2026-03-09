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
	auth := middleware.AuthMiddleware(h.userService)

	users.POST("/register", h.Register)
	users.POST("/login", h.Login)

	internal := router.Group("/internal/users")
	internal.POST("/verify-pin", h.InternalVerifyPIN)
	internal.GET("/lookup", h.InternalLookupByEmail)

	users.GET("/stats", auth, h.AdminGetStats)

	users.GET("", auth, h.ListUsers)
	users.GET("/", auth, h.ListUsers)

	users.GET("/profile", auth, h.GetProfile)
	users.PUT("/pin", auth, h.SetPin)
	users.PUT("/kyc", auth, h.UpdateKYC)
	users.POST("/logout", auth, h.Logout)

	users.PUT("/:id/kyc/verify", auth, h.AdminVerifyKYC)
}

func (h *UserHandler) AdminGetStats(c *gin.Context) {
	total, verified, err := h.userService.AdminGetStats(c.Request.Context())
	if err != nil {
		log.Printf("[ERROR] AdminGetStats: %v", err)
		response.InternalServerError(c, "failed to retrieve stats")
		return
	}
	response.OK(c, "stats retrieved", gin.H{
		"total_users":    total,
		"verified_users": verified,
	})
}

func (h *UserHandler) InternalLookupByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		response.BadRequest(c, "email query parameter is required", nil)
		return
	}

	user, err := h.userService.LookupByEmail(c.Request.Context(), email)
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, "lookup failed")
		return
	}

	c.JSON(200, gin.H{"success": true, "data": gin.H{"user_id": user.ID, "name": user.Name}})
}

func (h *UserHandler) InternalVerifyPIN(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		PIN    string `json:"pin"    binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}

	if err := h.userService.VerifyPIN(c.Request.Context(), req.UserID, req.PIN); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		c.JSON(200, gin.H{"success": false, "message": "invalid transaction PIN"})
		return
	}

	response.OK(c, "PIN verified", nil)
}

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

func (h *UserHandler) Logout(c *gin.Context) {
	userID, _ := c.Get(middleware.AuthUserIDKey)

	if err := h.userService.Logout(c.Request.Context(), userID.(string)); err != nil {
		log.Printf("[ERROR] Logout: %v", err)
		response.InternalServerError(c, "failed to logout")
		return
	}

	response.OK(c, "logout successful", nil)
}

func (h *UserHandler) AdminVerifyKYC(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		response.BadRequest(c, "user id is required", nil)
		return
	}

	if err := h.userService.AdminVerifyKYC(c.Request.Context(), userID); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		log.Printf("[ERROR] AdminVerifyKYC: %v", err)
		response.InternalServerError(c, "failed to verify KYC")
		return
	}

	response.OK(c, "KYC verified successfully", nil)
}

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
