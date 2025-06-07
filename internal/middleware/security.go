package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// SQLInjectionConfig holds configuration for SQL injection prevention
type SQLInjectionConfig struct {
	// Add blocked patterns that might indicate SQL injection
	BlockedPatterns []string
}

// NewSQLInjectionMiddleware creates a new middleware for SQL injection prevention
func NewSQLInjectionMiddleware() gin.HandlerFunc {
	// Common SQL injection patterns
	blockedPatterns := []string{
		`(?i)(select|update|delete|insert|drop|alter|create|truncate).*from`,
		`(?i)union.*select`,
		`(?i)into.*outfile`,
		`(?i)load_file`,
		`(?i)--`,
		`(?i)/\*`,
		`(?i)\*/`,
		`(?i);`,
	}

	compiledPatterns := make([]*regexp.Regexp, len(blockedPatterns))
	for i, pattern := range blockedPatterns {
		compiledPatterns[i] = regexp.MustCompile(pattern)
	}

	return func(c *gin.Context) {
		// Check query parameters
		for _, param := range c.Request.URL.Query() {
			for _, value := range param {
				if containsSQLInjection(value, compiledPatterns) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "Possible SQL injection detected",
					})
					return
				}
			}
		}

		// Check request body if it's a form
		if strings.Contains(c.GetHeader("Content-Type"), "application/x-www-form-urlencoded") {
			if err := c.Request.ParseForm(); err == nil {
				for _, values := range c.Request.Form {
					for _, value := range values {
						if containsSQLInjection(value, compiledPatterns) {
							c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
								"error": "Possible SQL injection detected",
							})
							return
						}
					}
				}
			}
		}

		c.Next()
	}
}

func containsSQLInjection(value string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(value) {
			return true
		}
	}
	return false
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// X-Content-Type-Options prevents MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options prevents clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Strict-Transport-Security enforces HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// X-XSS-Protection enables browser's XSS filter
		c.Header("X-XSS-Protection", "1; mode=block")

		// Content-Security-Policy prevents XSS, clickjacking and other injections
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; frame-ancestors 'none';")

		// Referrer-Policy controls how much referrer information should be included
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy controls which features and APIs can be used
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
