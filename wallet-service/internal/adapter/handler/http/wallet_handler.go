package http

import (
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/omni-wallet/wallet-service/internal/adapter/handler/http/middleware"
	"github.com/omni-wallet/wallet-service/internal/adapter/handler/http/response"
	"github.com/omni-wallet/wallet-service/internal/core/domain"
	"github.com/omni-wallet/wallet-service/internal/core/services"
)

// WalletHandler handles all HTTP requests for the Wallet Service.
// It validates input and delegates business logic to the appropriate service.
type WalletHandler struct {
	walletService   *services.WalletService
	transferService *services.TransferService
	validate        *validator.Validate
	jwtSecret       string
}

func NewWalletHandler(
	walletService *services.WalletService,
	transferService *services.TransferService,
	jwtSecret string,
) *WalletHandler {
	return &WalletHandler{
		walletService:   walletService,
		transferService: transferService,
		validate:        validator.New(),
		jwtSecret:       jwtSecret,
	}
}

// RegisterRoutes mounts all wallet-related routes onto the provided router group.
func (h *WalletHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := middleware.AuthMiddleware(h.jwtSecret)

	wallets := router.Group("/wallets")
	{
		// This endpoint is called internally by the User Service after registration.
		// In a production system it would be protected by an internal service token.
		wallets.POST("", h.CreateWallet)

		protected := wallets.Group("")
		protected.Use(auth)
		{
			protected.GET("/balance", h.GetBalance)
			protected.GET("/mutations", h.GetMutations)
			protected.GET("/transactions", h.GetTransactionHistory)
		}
	}

	transfers := router.Group("/transfers")
	transfers.Use(auth)
	{
		// Top-up is invoked by the mock VA webhook (no auth required in production,
		// but we keep auth here for testing; a real webhook would use HMAC validation).
		transfers.POST("/topup", h.Topup)
		transfers.POST("/p2p", h.Transfer)
	}
}

// CreateWallet godoc
// @Summary      Create a wallet for an existing user
// @Tags         wallets
// @Accept       json
// @Produce      json
// @Param        body  body      domain.CreateWalletRequest  true  "Create Wallet payload"
// @Success      201   {object}  response.APIResponse
// @Router       /wallets [post]
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req domain.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	wallet, err := h.walletService.CreateWallet(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, services.ErrWalletAlreadyExists) {
			response.Conflict(c, err.Error())
			return
		}
		log.Printf("[ERROR] CreateWallet: %v", err)
		response.InternalServerError(c, "failed to create wallet")
		return
	}

	response.Created(c, "wallet created successfully", wallet)
}

// GetBalance godoc
// @Summary      Get the authenticated user's wallet balance
// @Tags         wallets
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.APIResponse
// @Router       /wallets/balance [get]
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userID := c.GetString(middleware.AuthUserIDKey)

	balance, err := h.walletService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrWalletNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		log.Printf("[ERROR] GetBalance: %v", err)
		response.InternalServerError(c, "failed to retrieve balance")
		return
	}

	response.OK(c, "balance retrieved successfully", balance)
}

// GetMutations godoc
// @Summary      Get paginated wallet mutation (ledger) history
// @Tags         wallets
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number (default: 1)"
// @Param        limit  query  int  false  "Items per page (default: 20)"
// @Success      200    {object}  response.APIResponse
// @Router       /wallets/mutations [get]
func (h *WalletHandler) GetMutations(c *gin.Context) {
	userID := c.GetString(middleware.AuthUserIDKey)
	page := queryIntOrDefault(c, "page", 1)
	limit := queryIntOrDefault(c, "limit", 20)

	result, err := h.walletService.GetMutations(c.Request.Context(), userID, page, limit)
	if err != nil {
		if errors.Is(err, services.ErrWalletNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		log.Printf("[ERROR] GetMutations: %v", err)
		response.InternalServerError(c, "failed to retrieve mutations")
		return
	}

	response.OK(c, "mutations retrieved successfully", result)
}

// GetTransactionHistory godoc
// @Summary      Get paginated transaction history
// @Tags         wallets
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number (default: 1)"
// @Param        limit  query  int  false  "Items per page (default: 20)"
// @Success      200    {object}  response.APIResponse
// @Router       /wallets/transactions [get]
func (h *WalletHandler) GetTransactionHistory(c *gin.Context) {
	userID := c.GetString(middleware.AuthUserIDKey)
	page := queryIntOrDefault(c, "page", 1)
	limit := queryIntOrDefault(c, "limit", 20)

	result, err := h.walletService.GetTransactionHistory(c.Request.Context(), userID, page, limit)
	if err != nil {
		if errors.Is(err, services.ErrWalletNotFound) {
			response.NotFound(c, err.Error())
			return
		}
		log.Printf("[ERROR] GetTransactionHistory: %v", err)
		response.InternalServerError(c, "failed to retrieve transaction history")
		return
	}

	response.OK(c, "transactions retrieved successfully", result)
}

// Topup godoc
// @Summary      Top-up wallet via Virtual Account webhook
// @Tags         transfers
// @Accept       json
// @Produce      json
// @Param        body  body      domain.TopupRequest  true  "Top-up payload"
// @Success      200   {object}  response.APIResponse
// @Router       /transfers/topup [post]
func (h *WalletHandler) Topup(c *gin.Context) {
	var req domain.TopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	tx, err := h.transferService.Topup(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrWalletNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, services.ErrWalletFrozen):
			response.UnprocessableEntity(c, err.Error())
		case errors.Is(err, services.ErrLockAcquireFailed):
			response.UnprocessableEntity(c, err.Error())
		default:
			log.Printf("[ERROR] Topup: %v", err)
			response.InternalServerError(c, "top-up failed")
		}
		return
	}

	response.OK(c, "top-up successful", tx)
}

// Transfer godoc
// @Summary      P2P transfer between two users
// @Tags         transfers
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body      domain.TransferRequest  true  "Transfer payload"
// @Success      200   {object}  response.APIResponse
// @Router       /transfers/p2p [post]
func (h *WalletHandler) Transfer(c *gin.Context) {
	var req domain.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	// Enforce that the authenticated user is the one initiating the transfer.
	authedUserID := c.GetString(middleware.AuthUserIDKey)
	if req.SourceUserID != authedUserID {
		response.Unauthorized(c, "source_user_id must match the authenticated user")
		return
	}

	tx, err := h.transferService.Transfer(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInsufficientBalance):
			response.UnprocessableEntity(c, err.Error())
		case errors.Is(err, services.ErrInvalidPIN):
			response.Unauthorized(c, err.Error())
		case errors.Is(err, services.ErrSameWallet):
			response.BadRequest(c, err.Error(), nil)
		case errors.Is(err, services.ErrWalletNotFound):
			response.NotFound(c, err.Error())
		case errors.Is(err, services.ErrWalletFrozen):
			response.UnprocessableEntity(c, err.Error())
		case errors.Is(err, services.ErrLockAcquireFailed):
			response.UnprocessableEntity(c, err.Error())
		default:
			log.Printf("[ERROR] Transfer: %v", err)
			response.InternalServerError(c, "transfer failed")
		}
		return
	}

	response.OK(c, "transfer successful", tx)
}

// formatValidationErrors converts validator.ValidationErrors to a human-readable map.
func formatValidationErrors(err error) map[string]string {
	fieldErrors := make(map[string]string)
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, fe := range validationErrors {
			fieldErrors[fe.Field()] = buildValidationMessage(fe)
		}
	}
	return fieldErrors
}

func buildValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "uuid4":
		return "must be a valid UUID v4"
	case "gt":
		return "must be greater than " + fe.Param()
	case "max":
		return "must not exceed " + fe.Param() + " characters"
	case "len":
		return "must be exactly " + fe.Param() + " characters"
	case "numeric":
		return "must contain only numeric digits"
	default:
		return "invalid value"
	}
}

// queryIntOrDefault parses a query parameter as int, returning defaultVal on failure.
func queryIntOrDefault(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val < 1 {
		return defaultVal
	}
	return val
}
