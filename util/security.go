package util

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type SecurityConfig struct {
	// Habilitar solamente cuando HTTPS esté garantizado.
	EnableHSTS bool

	// false = Content-Security-Policy-Report-Only
	// true  = Content-Security-Policy
	EnforceCSP bool

	// Duración HSTS en segundos.
	HSTSMaxAge int
}

func SecurityHeaders(cfg SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Evita MIME sniffing.
		c.Header(
			"X-Content-Type-Options",
			"nosniff",
		)

		// Protección contra Clickjacking.
		c.Header(
			"X-Frame-Options",
			"DENY",
		)

		// Protección de información de navegación.
		c.Header(
			"Referrer-Policy",
			"strict-origin-when-cross-origin",
		)

		// Restringe funcionalidades del navegador
		// que TourManager no necesita globalmente.
		c.Header(
			"Permissions-Policy",
			"camera=(), microphone=(), geolocation=(self)",
		)

		// HSTS: solo cuando HTTPS está garantizado.
		if cfg.EnableHSTS {
			maxAge := cfg.HSTSMaxAge
			if maxAge <= 0 {
				maxAge = 86400
			}

			c.Header(
				"Strict-Transport-Security",
				"max-age="+
					strconv.Itoa(maxAge),
			)
		}

		// CSP temporal en modo observación.
		policy := "default-src 'self'; " +
			"base-uri 'self'; " +
			"object-src 'none'; " +
			"frame-ancestors 'none'; " +
			"form-action 'self'; " +
			"script-src 'self'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: blob:; " +
			"font-src 'self' data:; " +
			"connect-src 'self'; " +
			"manifest-src 'self'; " +
			"worker-src 'self' blob:; " +
			"upgrade-insecure-requests"

		header := "Content-Security-Policy-Report-Only"

		if cfg.EnforceCSP {
			header = "Content-Security-Policy"
		}

		c.Header(header, policy)

		c.Next()
	}
}

// Generador opcional de nonce para una futura CSP estricta.
/*
func GenerateNonce() (string, error) {
    bytes := make([]byte, 32)

    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }

    return base64.StdEncoding.EncodeToString(bytes), nil
}
*/
