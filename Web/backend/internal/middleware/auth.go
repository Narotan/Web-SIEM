package middleware

import (
	"net/http"

	"github.com/Narotan/Web-SIEM/internal/config"
	"github.com/gin-gonic/gin"
)

// BasicAuth проверяет логин и пароль из конфига для каждого запроса
func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.GetConfig()

		// Получаем учетные данные из заголовка запроса
		user, password, hasAuth := c.Request.BasicAuth()

		// Проверяем соответствие учетным данным из переменных окружения
		if !hasAuth || user != cfg.WebUser || password != cfg.WebPass {
			// Если неверно — требуем авторизацию
			c.Header("WWW-Authenticate", `Basic realm="Authorization Required"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: invalid login or password",
			})
			return
		}

		// Если всё верно — продолжаем выполнение следующего обработчика
		c.Next()
	}
}
