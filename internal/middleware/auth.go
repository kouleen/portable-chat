package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/pkg/storecli"
	utils "github.com/kouleen/portable-chat/pkg/util"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			accessToken := c.Query("access_token")
			if accessToken == "" {
				utils.Error(c, http.StatusUnauthorized, "Unauthorized")
				c.Abort()
				return
			}
			authHeader = "Bearer " + accessToken
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, http.StatusUnauthorized, "Invalid Token")
			c.Abort()
			return
		}
		tokenString := parts[1]
		userHeaderJson, err := storecli.Get(tokenString)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		if userHeaderJson == "" {
			utils.Error(c, http.StatusUnauthorized, "Invalid Token")
			c.Abort()
			return
		}
		var authorization model.Authorization
		if err := json.Unmarshal([]byte(userHeaderJson), &authorization); err != nil {
			utils.Error(c, http.StatusInternalServerError, "Unmarshal err")
			c.Abort()
			return
		}
		ctx := utils.SetAuthorization(c.Request.Context(), &authorization)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
