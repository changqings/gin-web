package router

import (
	"github.com/changqings/gin-web/cicd"
	"github.com/gin-gonic/gin"
)

func CICDRouter(app *gin.Engine) {

	build := &cicd.DockerBuild{}
	deploy := &cicd.K8sDeploy{}
	cicdGroup := app.Group("/cicd")

	// build body josn
	// {
	// "project_name": "go-micro",
	// "group": "backend",
	// "repo_ssh": "ssh://git@gitlab.scq.com:522/backend/go-micro.git",
	// "project_type": "go",
	// "project_env": "dev",
	// "tag_or_branch": "v0.0.1"
	// }
	cicdGroup.POST("/build", build.Build())

	// post body json
	// {
	// "app_name": "go-micro",
	// "group": "backend",
	// "namespace": "default",
	// "project_type": "go",
	// "project_env": "dev",
	// "tag_or_branch": "v0.0.1"
	// }
	cicdGroup.POST("/deploy", deploy.Deploy())

}
