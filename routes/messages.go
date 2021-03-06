package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/ovh/tat/controllers"
)

// InitRoutesMessages initialized routes for Messages Controller
func InitRoutesMessages(router *gin.Engine) {
	messagesCtrl := &controllers.MessagesController{}

	g := router.Group("/messages")
	g.Use(CheckPassword())
	{
		g.GET("/*topic", messagesCtrl.List)
	}

	r := router.Group("/read")
	r.Use()
	{
		r.GET("/*topic", messagesCtrl.List)
	}

	gm := router.Group("/message")
	gm.Use(CheckPassword())
	{
		//Create a message, a reply, a bookmark
		gm.POST("/*topic", messagesCtrl.Create)

		// Like, Unlike, Label, Unlabel a message, mark as task
		gm.PUT("/*topic", messagesCtrl.Update)

		// Delete a bookmark
		gm.DELETE("/:idMessage", messagesCtrl.Delete)

	}

}
