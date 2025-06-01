package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
)

// NoAuthMiddleware - Middleware tạm thời để bỏ qua xác thực
// CẢNH BÁO: Chỉ sử dụng trong môi trường phát triển!
func NoAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Tạo fake user claims để tránh lỗi khi code cần thông tin user
		fakeUser := models.JWTClaims{
			UserID:    1,
			Username:  "temp_user",
			Role:      "admin",
			CompanyID: 1,
		}
		
		// Set fake user data vào context
		c.Locals("userID", fakeUser.UserID)
		c.Locals("username", fakeUser.Username)
		c.Locals("role", fakeUser.Role)
		c.Locals("companyID", fakeUser.CompanyID)
		c.Locals("user", fakeUser)
		
		return c.Next()
	}
}

// NoRoleMiddleware - Middleware tạm thời để bỏ qua kiểm tra role
// CẢNH BÁO: Chỉ sử dụng trong môi trường phát triển!
func NoRoleMiddleware(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Bỏ qua tất cả kiểm tra role
		return c.Next()
	}
}
