package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

// newReverseProxy creates an httputil.ReverseProxy that forwards all requests
// to the given targetURL, preserving the original request path and query string.
// A reasonable timeout is configured on the transport layer.
func newReverseProxy(targetURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid upstream URL %q: %w", targetURL, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Use a transport with explicit timeouts to prevent goroutine leaks when
	// an upstream service is slow or unresponsive.
	proxy.Transport = &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		DisableCompression:    true,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	// Replace the default error handler so upstream failures return structured JSON.
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(rw, `{"success":false,"message":"upstream service unavailable: %s"}`, err.Error())
	}

	return proxy, nil
}

// proxyHandlerFunc returns a Gin handler that proxies the request to the given proxy.
func proxyHandlerFunc(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

// publicPaths is the set of "method:path" combinations that bypass JWT auth.
// Using the real request URL path (not the Gin route pattern) avoids the
// static-segment vs catch-all wildcard conflict that Gin's radix tree enforces.
var publicPaths = map[string]bool{
	"POST:/api/v1/users/register": true,
	"POST:/api/v1/users/login":    true,
}

// RegisterRoutes sets up all proxy routes on the Gin engine.
//
// Route table (all paths use a single wildcard catch-all per upstream):
//
//	* /api/v1/users/**     → user-service   (POST /register and /login are public;
//	                                          all other user routes require JWT)
//	* /api/v1/wallets/**   → wallet-service (JWT required)
//	* /api/v1/transfers/** → wallet-service (JWT required)
//
// Why single catch-all instead of static + wildcard:
// Gin's radix-tree router panics at startup when a static segment (e.g.
// /api/v1/users/register) and a catch-all wildcard (/api/v1/users/*path) share
// the same prefix. The fix is to register only the wildcard and handle the
// public/protected distinction in middleware via the real URL path.
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

	// Apply rate limiter globally to every route.
	engine.Use(rateLimiter)

	// ------------------------------------------------------------------
	// User Service — single catch-all wildcard.
	// The selectiveAuth middleware skips JWT for public paths defined in
	// publicPaths and enforces it for everything else.
	// ------------------------------------------------------------------
	selective := selectiveAuthMiddleware(authMiddleware)

	userGroup := engine.Group("/api/v1/users", selective)
	{
		userGroup.Any("/*path", userHandler)
	}

	// ------------------------------------------------------------------
	// Wallet Service — all routes require JWT.
	// ------------------------------------------------------------------
	protected := engine.Group("/", authMiddleware)
	{
		protected.Any("/api/v1/wallets/*path", walletHandler)
		protected.Any("/api/v1/transfers/*path", walletHandler)
	}

	return nil
}

// selectiveAuthMiddleware wraps the given auth middleware and skips it when
// the real request path (method + URL.Path) is in the publicPaths whitelist.
// This avoids registering separate static routes alongside a catch-all wildcard.
func selectiveAuthMiddleware(auth gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.Request.Method + ":" + ctx.Request.URL.Path
		if publicPaths[key] {
			// Public path — skip JWT verification entirely.
			ctx.Next()
			return
		}
		// Protected path — delegate to the real auth middleware.
		auth(ctx)
	}
}
