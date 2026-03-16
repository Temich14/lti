package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type LaunchAdapter struct {
	ltiService LTIService
}

func NewLaunchAdapter(ltiService LTIService) *LaunchAdapter {
	return &LaunchAdapter{ltiService: ltiService}
}

type LaunchRequest struct {
	IdToken string `json:"id_token"`
	State   string `json:"state"`
}

func (l *LaunchAdapter) Launch(c *gin.Context) {
	var req LaunchRequest
	req.IdToken = c.PostForm("id_token")
	req.State = c.PostForm("state")

	members, err := l.ltiService.Launch(c.Request.Context(), req.IdToken, req.State)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = members

	c.HTML(http.StatusOK, "deeplinking/select.html", gin.H{
		"title":          "Select content to deeplink",
		"text":           "Choose content items to return to the platform.",
		"acceptMultiple": true,
		"acceptTypes":    []string{"ltiResourceLink"},
		"returnUrl":      c.Query("return_url"),
	})
}
