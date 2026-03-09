package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	xenditpay "github.com/omni-wallet/wallet-service/internal/adapter/payment/xendit"
	"github.com/omni-wallet/wallet-service/internal/adapter/handler/http/middleware"
	"github.com/omni-wallet/wallet-service/internal/adapter/handler/http/response"
	"github.com/omni-wallet/wallet-service/internal/core/domain"
	"github.com/omni-wallet/wallet-service/internal/core/services"
)

type WalletHandler struct {
	walletService   *services.WalletService
	transferService *services.TransferService
	paymentService  *services.PaymentService
	validate        *validator.Validate
	jwtSecret       string
}

func NewWalletHandler(
	walletService *services.WalletService,
	transferService *services.TransferService,
	paymentService *services.PaymentService,
	jwtSecret string,
) *WalletHandler {
	return &WalletHandler{
		walletService:   walletService,
		transferService: transferService,
		paymentService:  paymentService,
		validate:        validator.New(),
		jwtSecret:       jwtSecret,
	}
}

func (h *WalletHandler) RegisterRoutes(router *gin.RouterGroup) {
	auth := middleware.AuthMiddleware(h.jwtSecret)

	wallets := router.Group("/wallets")
	{
		wallets.POST("", h.CreateWallet)

		protected := wallets.Group("")
		protected.Use(auth)
		{
			protected.GET("/balance", h.GetBalance)
			protected.GET("/mutations", h.GetMutations)
			protected.GET("/transactions", h.GetTransactionHistory)
			protected.GET("/stats", h.AdminGetStats)
		}
	}

	transfers := router.Group("/transfers")
	transfers.Use(auth)
	{
		transfers.POST("/topup", h.Topup)
		transfers.POST("/p2p", h.Transfer)
		transfers.POST("/topup/va", h.RequestVA)
	}

	payments := router.Group("/payments/xendit")
	{
		payments.POST("/callback", h.XenditCallback)
		if os.Getenv("APP_ENV") != "production" {
			payments.POST("/simulate", auth, h.SimulateXenditPayment)
		}
	}
}

func (h *WalletHandler) AdminGetStats(c *gin.Context) {
	totalTx, totalVolume, err := h.walletService.AdminGetStats(c.Request.Context())
	if err != nil {
		log.Printf("[ERROR] AdminGetStats: %v", err)
		response.InternalServerError(c, "failed to retrieve stats")
		return
	}
	response.OK(c, "stats retrieved", gin.H{
		"total_transactions": totalTx,
		"total_volume":       totalVolume,
	})
}

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

func (h *WalletHandler) RequestVA(c *gin.Context) {
	if h.paymentService == nil {
		response.BadRequest(c, "payment gateway not configured", nil)
		return
	}

	var req domain.RequestVARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body", err.Error())
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(c, "validation failed", formatValidationErrors(err))
		return
	}

	userID := c.GetString(middleware.AuthUserIDKey)
	va, err := h.paymentService.GetOrCreateVA(c.Request.Context(), userID, req.Name, req.BankCode)
	if err != nil {
		log.Printf("[ERROR] RequestVA user=%s: %v", userID, err)

		var xenditErr *xenditpay.APIError
		if errors.As(err, &xenditErr) {
			var msg string
			switch xenditErr.Code {
			case "REQUEST_FORBIDDEN_ERROR":
				msg = "API key Xendit tidak memiliki izin untuk membuat Virtual Account. Periksa permission di dashboard Xendit."
			case "DUPLICATE_CALLBACK_VIRTUAL_ACCOUNT":
				msg = "Virtual Account sudah ada namun gagal dibaca dari cache. Coba lagi."
			default:
				msg = fmt.Sprintf("Payment gateway error [%s]: %s", xenditErr.Code, xenditErr.Message)
			}
			c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": msg})
			return
		}

		response.InternalServerError(c, "gagal membuat virtual account")
		return
	}

	response.OK(c, "virtual account berhasil dibuat", va)
}

func (h *WalletHandler) XenditCallback(c *gin.Context) {
	if h.paymentService == nil {
		c.Status(200)
		log.Println("[WARN] XenditCallback: payment service not configured, ignoring")
		return
	}

	token := c.GetHeader("x-callback-token")
	if !h.paymentService.Gateway().VerifyWebhookToken(token) {
		log.Printf("[WARN] XenditCallback: invalid callback token")
		c.JSON(401, gin.H{"message": "unauthorized"})
		return
	}

	var payment domain.XenditVAPayment
	if err := c.ShouldBindJSON(&payment); err != nil {
		response.BadRequest(c, "invalid payload", err.Error())
		return
	}

	tx, err := h.paymentService.ProcessXenditPayment(c.Request.Context(), payment)
	if err != nil {
		log.Printf("[ERROR] XenditCallback payment=%s: %v", payment.PaymentID, err)
		response.InternalServerError(c, "payment processing failed")
		return
	}

	response.OK(c, "payment processed", tx)
}

func (h *WalletHandler) SimulateXenditPayment(c *gin.Context) {
	if h.paymentService == nil {
		response.BadRequest(c, "payment service not configured", nil)
		return
	}

	var payment domain.XenditVAPayment
	if err := c.ShouldBindJSON(&payment); err != nil {
		response.BadRequest(c, "invalid payload", err.Error())
		return
	}

	tx, err := h.paymentService.ProcessXenditPayment(c.Request.Context(), payment)
	if err != nil {
		log.Printf("[ERROR] SimulateXenditPayment: %v", err)
		response.InternalServerError(c, "simulation failed")
		return
	}

	response.OK(c, "simulated payment processed", tx)
}

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
		case errors.Is(err, services.ErrTargetUserNotFound):
			response.NotFound(c, "target user not found")
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
