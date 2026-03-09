package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

func newReverseProxy(targetURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL %q: %w", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    true,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(rw, `{"success":false,"message":"upstream service unavailable: %s"}`, err.Error())
	}

	return proxy, nil
}

func proxyHandlerFunc(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

var publicPaths = map[string]bool{
	"POST:/api/v1/users/register":              true,
	"POST:/api/v1/users/login":                 true,
	"POST:/api/v1/payments/xendit/callback":    true,
}

func RegisterRoutes(
	engine *gin.Engine,
	authMiddleware gin.HandlerFunc,
	rateLimiter gin.HandlerFunc,
	userServiceURL string,
	walletServiceURL string,
) error {
	userProxy, err := newReverseProxy(userServiceURL)
	if err != nil {
		return err
	}

	walletProxy, err := newReverseProxy(walletServiceURL)
	if err != nil {
		return err
	}

	userHandler := proxyHandlerFunc(userProxy)
	walletHandler := proxyHandlerFunc(walletProxy)

	engine.Use(rateLimiter)

	selective := selectiveAuthMiddleware(authMiddleware)

	userGroup := engine.Group("/api/v1/users", selective)
	{
		userGroup.Any("", userHandler)
		userGroup.Any("/*path", userHandler)
	}

	protected := engine.Group("/", authMiddleware)
	{
		protected.Any("/api/v1/wallets/*path", walletHandler)
		protected.Any("/api/v1/transfers/*path", walletHandler)
	}

	paymentsSelective := selectiveAuthMiddleware(authMiddleware)
	paymentsGroup := engine.Group("/api/v1/payments", paymentsSelective)
	{
		paymentsGroup.Any("/*path", walletHandler)
	}

	return nil
}

func selectiveAuthMiddleware(auth gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Request.Method + ":" + ctx.Request.URL.Path
		if publicPaths[key] {
			ctx.Next()
			return
		}
		auth(ctx)
	}
}
