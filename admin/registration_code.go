package admin

import (
	"database/sql"
	"net/http"
	"strconv"

	"neoai/auth"
	"neoai/utils"

	"github.com/gin-gonic/gin"
)

// registration_code.go — admin endpoints for managing registration codes.
//
//   GET    /admin/registration-code/list      — list all codes
//   POST   /admin/registration-code/generate   — generate N new codes
//   GET    /admin/registration-code/state     — is registration code required?
//   POST   /admin/registration-code/state     — toggle require_code flag
//   POST   /admin/registration-code/disable   — disable a code
//   GET    /admin/registration-code/delete/:id — delete a code

type GenerateRegCodeForm struct {
	Number  int     `json:"number"`
	Quota   float32 `json:"quota"`
	MaxUses int     `json:"max_uses"`
	Note    string  `json:"note"`
}

type RegCodeStateForm struct {
	Required bool `json:"required"`
}

type RegCodeDisableForm struct {
	Id int64 `json:"id"`
}

func ListRegCodesAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	codes, err := auth.ListRegistrationCodes(db)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   codes,
	})
}

func GenerateRegCodesAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	var form GenerateRegCodeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	if form.Number <= 0 || form.Number > 1000 {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": "number must be between 1 and 1000",
		})
		return
	}
	if form.MaxUses <= 0 {
		form.MaxUses = 1
	}
	codes, err := auth.GenerateRegistrationCodes(db, form.Number, form.Quota, form.MaxUses, form.Note)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   codes,
	})
}

func GetRegCodeStateAPI(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   true,
		"required": auth.IsRegistrationCodeRequired(),
	})
}

func SetRegCodeStateAPI(c *gin.Context) {
	var form RegCodeStateForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	// Persist to config.yaml under system.registration.require_code
	if err := utils.SaveConfig("system.registration.require_code", form.Required); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   true,
		"required": form.Required,
	})
}

func DisableRegCodeAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	var form RegCodeDisableForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	if err := auth.DisableRegistrationCode(db, form.Id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}

func DeleteRegCodeAPI(c *gin.Context) {
	db := utils.GetDBFromContext(c)
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := auth.DeleteRegistrationCode(db, id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true})
}

// silence unused import warning (sql is referenced by auth indirectly)
var _ = sql.ErrNoRows
