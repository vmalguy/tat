package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/models"
	"github.com/ovh/tat/utils"
	"github.com/spf13/viper"
)

// private
var (
	tatHeaderPassword          = "Tat_password"
	tatHeaderPasswordLower     = "tat_password"
	tatHeaderPasswordLowerDash = "tat-password"
)

type tatHeaders struct {
	username      string
	password      string
	trustUsername string
}

// CheckAdmin is a middleware, abort request if user is not admin
func CheckAdmin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !utils.IsTatAdmin(ctx) {
			ctx.AbortWithError(http.StatusForbidden, errors.New("user is not admin"))
		}
	}
}

// CheckPassword is a middleware, check username / password in Request Header and validate
// them in DB. If username/password is invalid, abort request
func CheckPassword() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// refresh store to avoid lost connection on mongo
		models.RefreshStore()

		tatHeaders, err := extractTatHeaders(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		user, err := validateTatHeaders(tatHeaders)
		if err != nil {
			log.Errorf("Error, send 401, err : %s", err.Error())
			ctx.AbortWithError(http.StatusUnauthorized, err)
			return
		}

		err = storeInContext(ctx, user)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
}

// extractTatHeadesr extracts Tat_username and Tat_password from Headers Request
// try match tat_username, tat_password, tat-username, tat-password
// try dash version, thanks to perl lib...
func extractTatHeaders(ctx *gin.Context) (tatHeaders, error) {
	var tatHeaders tatHeaders

	for k, v := range ctx.Request.Header {
		if strings.ToLower(k) == utils.TatHeaderUsernameLower {
			tatHeaders.username = v[0]
		} else if strings.ToLower(k) == tatHeaderPasswordLower {
			tatHeaders.password = v[0]
		} else if strings.ToLower(k) == utils.TatHeaderUsernameLowerDash {
			tatHeaders.username = v[0]
		} else if strings.ToLower(k) == tatHeaderPasswordLowerDash {
			tatHeaders.password = v[0]
		} else if k == viper.GetString("header_trust_username") {
			tatHeaders.trustUsername = v[0]
		}
	}

	if tatHeaders.password != "" && tatHeaders.username != "" {
		return tatHeaders, nil
	}

	if tatHeaders.trustUsername != "" && tatHeaders.trustUsername != "null" {
		return tatHeaders, nil
	}

	return tatHeaders, errors.New("Invalid Tat Headers")
}

// validateTatHeaders fetch user in db and check Password
func validateTatHeaders(tatHeaders tatHeaders) (models.User, error) {

	user := models.User{}
	if tatHeaders.trustUsername != "" && tatHeaders.trustUsername != "null" {
		err := user.TrustUsername(tatHeaders.trustUsername)
		if err != nil {
			return user, fmt.Errorf("User %s does not exist. Please register before. Err:%s", tatHeaders.trustUsername, err.Error())
		}
	} else {
		err := user.FindByUsernameAndPassword(tatHeaders.username, tatHeaders.password)
		if err != nil {
			return user, fmt.Errorf("Invalid Tat credentials for username %s, err:%s", tatHeaders.username, err.Error())
		}
	}

	return user, nil
}

// storeInContext stores username and isAdmin flag only
func storeInContext(ctx *gin.Context, user models.User) error {
	ctx.Set(utils.TatHeaderUsername, user.Username)
	ctx.Set(utils.TatCtxIsAdmin, user.IsAdmin)
	ctx.Set(utils.TatCtxIsSystem, user.IsSystem)

	if user.IsAdmin {
		log.Debugf("user %s isAdmin", user.Username)
	}

	if user.IsSystem {
		log.Debugf("user %s isSystem", user.Username)
	}

	return nil
}
