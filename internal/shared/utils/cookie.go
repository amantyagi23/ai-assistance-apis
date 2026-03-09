package utils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func SetCookieHandler(c *fiber.Ctx, tokenName string, token string, expiration time.Time) {

	cookie := new(fiber.Cookie)
	cookie.Name = tokenName
	cookie.Value = token
	cookie.Expires = expiration
	cookie.Path = "/"
	cookie.HTTPOnly = true
	cookie.Secure = true
	cookie.SameSite = "Lax"

	c.Cookie(cookie)

	return
}
