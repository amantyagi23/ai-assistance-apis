package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetCookieHandler(c *fiber.Ctx, tokenName string, token string, expiration time.Time) error {

	cookie := new(fiber.Cookie)
	cookie.Name = tokenName
	cookie.Value = token
	cookie.Expires = expiration
	cookie.Path = "/"
	cookie.HTTPOnly = true
	cookie.Secure = false
	cookie.SameSite = "Lax"

	c.Cookie(cookie)
	return nil
}
