package cicd

import (
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// write this ci_cd just for learn use golang.

// run in docker and only on linux, code on ubuntu
// 过程：代码仓库(gitlab_on_docker) --> 挂载到特定容器(docker local) --> 构建并生成产物(use Dockerfile)
// --> 构建镜像并上传到镜像仓库(docker-registry)

// gitlab_host = http://192.168.1.15

// ssh://git@gitlab.scq.com:522/devops/go-ws.git

var (
	TimeLayOutSet           = "20060102150405.000"
	TxTcrHost               = "ccr.ccs.tencentyun.com"
	TxTcrNamespaceDevopsScq = "devops_scq"
	LocalBuildBaseDir       = "/data/devops/build/"
	GitlabCICDRepoAddr      = "ssh://git@gitlab.scq.com:522/devops/cicd.git"
	CICDRepoLocalPath       = "devops_cicd"
	buildConfig             = ConfigCICD{BuildHistoryReserve: 10}
)

type ConfigCICD struct {
	BuildHistoryReserve int
}

// use ~/.ssh/id_rsa as git repo private key
type DockerBuild struct {
	ProjectName      string `json:"project_name"`
	Group            string `json:"group"`
	RepoSSH          string `json:"repo_ssh"`
	ProjectLocalPath string `json:"-"`
	TimestampNowDir  string `json:"-"`
	WorkDir          string `json:"-"`
	Type             string `json:"project_type"`
	Env              string `json:"project_env"`
	Tag              string `json:"tag_or_branch"`
	Image            string `json:"-"`
}

// first use Dockerfile of devops/cicd project.Dockerfile_<buildEnv>
// second use Dockerfile of devops/cicd project.Dokerfile
// third use Dockerfile of repo.Dockerfile
func NewDockerBuild(name, group, sshAddr, buildType, buildEnv, tagOrbranch string) *DockerBuild {
	image := filepath.Join(TxTcrHost, TxTcrNamespaceDevopsScq, name+":"+buildEnv+"-"+tagOrbranch)
	localPath := filepath.Join(LocalBuildBaseDir, group, name)
	timeStampStr := "t_" + time.Now().Local().Format(TimeLayOutSet)
	workDir := filepath.Join(localPath, timeStampStr)

	return &DockerBuild{
		ProjectName:      name,
		Group:            group,
		ProjectLocalPath: localPath, // this for delete old clone
		RepoSSH:          sshAddr,
		WorkDir:          workDir,
		Type:             buildType,
		Env:              buildEnv,
		Tag:              tagOrbranch,
		Image:            image,
		TimestampNowDir:  timeStampStr,
	}
}

func (d *DockerBuild) SetAllPath() {
	image := filepath.Join(TxTcrHost, TxTcrNamespaceDevopsScq, d.ProjectName+":"+d.Env+"-"+d.Tag)
	localPath := filepath.Join(LocalBuildBaseDir, d.Group, d.ProjectName)
	timeStampStr := "t_" + time.Now().Local().Format(TimeLayOutSet)
	workDir := filepath.Join(localPath, timeStampStr)

	d.Image = image
	d.ProjectLocalPath = localPath
	d.WorkDir = workDir

}

// clone
func (d *DockerBuild) DoClone(c ConfigCICD) error {

	err := os.MkdirAll(d.WorkDir, 0755)
	if err != nil {
		slog.Error("build mkdir all", "path", d.WorkDir, "msg", err)
		return err
	}

	//clone 之用先清理过多的目录
	fss, err := os.ReadDir(d.ProjectLocalPath)
	if err != nil {
		return err
	}

	if len(fss) > c.BuildHistoryReserve &&
		strings.Contains(d.WorkDir, "t_") &&
		strings.HasPrefix(d.WorkDir, LocalBuildBaseDir) {
		//do clean
		for _, f := range fss[:len(fss)-c.BuildHistoryReserve] {
			err := os.RemoveAll(filepath.Join(d.ProjectLocalPath, f.Name()))
			if err != nil {
				slog.Error("remove dir", "name", f.Name(), "msg", err)
				continue
			}
			slog.Info("remove dir", "path", f.Name(), "msg", "success")
		}
	}

	cloneApp := NewCmd(
		"git", "clone",
		"-b", d.Tag,
		"--depth=1",
		d.RepoSSH, d.WorkDir)

	cloneCICD := NewCmd(
		"git", "clone",
		"-b", "main",
		"--depth=1",
		GitlabCICDRepoAddr, filepath.Join(d.WorkDir, CICDRepoLocalPath),
	)

	// debug cmd
	// fmt.Printf("cmd.String()=%s\n", cloneApp.String())

	if err := cloneApp.Run(); err != nil {
		return err
	}
	if err := cloneCICD.Run(); err != nil {
		return err
	}

	return nil
}

// build

func (d *DockerBuild) DoBuild() error {

	// Dockerfile define
	var dockerFileName string
	dockerfileNameDefault := "Dockerfile"
	dockerfileNameWithEnv := "Dockerfile_" + d.Env

	// workDir and devops/cicd build config
	devopsCICDBuildPath := filepath.Join(d.WorkDir, CICDRepoLocalPath, "jobs", d.Group, d.ProjectName, "build")

	// first use devops/cicd build
	var fss []string
	files, err := os.ReadDir(devopsCICDBuildPath)
	if err != nil {
		slog.Warn("build use devops/cicd dir", "path", devopsCICDBuildPath, "msg", err)
	} else {
		for _, f := range files {
			fss = append(fss, f.Name())
		}
	}
	if len(fss) == 0 {
		dockerFileName = dockerfileNameDefault
		slog.Info("build use dockerfile", "msg", "repo.Dockerfile")
	} else {
		// cp all file of build to project localpath
		for _, f := range fss {
			copyCmd := NewCmd("cp", "-a", filepath.Join(devopsCICDBuildPath, f), d.WorkDir)
			err := copyCmd.Run()
			if err != nil {
				return err
			}
		}

		if slices.Contains(fss, dockerfileNameWithEnv) {
			dockerFileName = dockerfileNameWithEnv
			slog.Info("build use dockerfile", "msg", "devops_cicd."+dockerFileName)
		} else {
			dockerFileName = dockerfileNameDefault
			slog.Info("build use dockerfile", "msg", "devops_cicd."+dockerFileName)
		}
	}

	// run docker build
	cmd := NewCmd("docker", "build",
		"-t", d.Image,
		"-f", dockerFileName,
		".",
	)
	cmd.Dir = d.WorkDir

	return cmd.Run()

}

// push
func (d *DockerBuild) DoPush() error {

	pushCmd := NewCmd("docker", "push", d.Image)
	return pushCmd.Run()
}

// gin handlerfunc

func (d *DockerBuild) Build() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		// bind
		if err := ctx.ShouldBind(&d); err != nil {
			slog.Error("build bind to &DockerBuild{}", "msg", err)
			ctx.AbortWithError(482, errors.New("post body json not right for build, please check"))
			return
		}

		// setAllPath
		d.SetAllPath()

		// clone
		errClone := d.DoClone(buildConfig)
		if errClone != nil {
			slog.Error("main build clone", "git_repo", d.RepoSSH, "msg", errClone)
			ctx.AbortWithError(482, errors.New("clone failed"))
			return
		}

		// build
		errBuild := d.DoBuild()
		if errBuild != nil {
			slog.Error("main docker build", "image", d.Image, "msg", errBuild)
			ctx.AbortWithError(482, errors.New("docker build failed"))
			return
		}

		// push
		errPush := d.DoPush()
		if errPush != nil {
			slog.Error("main docker push", "image", d.Image, "msg", errPush)
			ctx.AbortWithError(482, errors.New("docker push failed"))
			return
		}

		ctx.JSON(200, gin.H{
			"project_name":  d.ProjectName,
			"project_group": d.Group,
			"project_env":   d.Env,
			"push_image":    d.Image,
			"msg":           "build and push success.",
		})

	}

}

// cmd with os.Strerr and os.Stdout
func NewCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	// default stderr stdout to /dev/null
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd

}
