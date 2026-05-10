package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type launchFlowRunner interface {
	Run(ctx context.Context, idToken, state string) (deeplinkRelativeURL string, err error)
}

type LaunchAdapter struct {
	flow launchFlowRunner
}

func NewLaunchAdapter(flow launchFlowRunner) *LaunchAdapter {
	return &LaunchAdapter{flow: flow}
}

func (l *LaunchAdapter) Launch(c *gin.Context) {
	idToken := c.PostForm("id_token")
	state := c.PostForm("state")

	redirectPath, err := l.flow.Run(c.Request.Context(), idToken, state)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusSeeOther, redirectPath)
}
