package ci

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// write this ci_cd just for learn use golang.

// run in docker and only on linux, code on ubuntu
// 过程：代码仓库(gitlab_on_docker) --> 挂载到特定容器(docker local) --> 构建并生成产物(use Dockerfile)
// --> 构建镜像并上传到镜像仓库(docker-registry)

// gitlab_host = http://192.168.1.15

// ssh://git@gitlab.scq.com:522/devops/go-ws.git

var (
	LayOutSet                   = "20060102150405.000"
	TX_TCR_HOST                 = "ccr.ccs.tencentyun.com"
	TX_TCR_NAMESPACE_DEVOPS_SCQ = "devops_scq"
	LOCALPATH_PARATENT_DIR      = "/data/build/gitlab/"
	GITLAB_CICD_REPO_ADDR       = "ssh://git@gitlab.scq.com:522/devops/cicd.git"
	CICD_REPO_LOCAL_PATH        = "devops_cicd"
)

type ConfigCICD struct {
	BuildHistoryReserve int
}

// use ~/.ssh/id_rsa as git repo private key
type GitlabRepoClone struct {
	ProjectName      string
	Group            string
	RepoSSH          string
	TagOrBranch      string
	ProjecLocaltPath string
	TimestampNowDir  string
}

type DockerBuild struct {
	ProjectName      string
	ProjectLocalPath string
	Type             string
	Env              string
	Tag              string
	Image            string
}

func (d *DockerBuild) DoPush() error {

	pushCmd := NewCmd("docker", "push", d.Image)
	return pushCmd.Run()
}

func (d *DockerBuild) DoBuild(g *GitlabRepoClone) error {

	// Dockerfile define
	var dockerFileName string
	dockerfileNameDefault := "Dockerfile"
	dockerfileNameWithEnv := "Dockerfile_" + d.Env

	// workDir and devops/cicd build config
	workDir := filepath.Join(g.ProjecLocaltPath, g.TimestampNowDir)
	devopsCICDBuildPath := filepath.Join(workDir, CICD_REPO_LOCAL_PATH, "jobs", g.Group, d.ProjectName, "build")

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
			copyCmd := NewCmd("cp", "-a", filepath.Join(devopsCICDBuildPath, f), workDir)
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
	cmd.Dir = workDir

	return cmd.Run()

}

// first use Dockerfile of devops/cicd project.Dockerfile_<buildEnv>
// second use Dockerfile of devops/cicd project.Dokerfile
// third use Dockerfile of repo.Dockerfile
func NewDockerBuild(name, path, buildType, buildTag, buildEnv string) *DockerBuild {
	image := filepath.Join(TX_TCR_HOST, TX_TCR_NAMESPACE_DEVOPS_SCQ, name+":"+buildEnv+"-"+buildTag)
	return &DockerBuild{
		ProjectName:      name,
		ProjectLocalPath: path,
		Type:             buildType,
		Env:              buildEnv,
		Tag:              buildTag,
		Image:            image,
	}
}

func NewGitlabRepoClone(group, name, sshAddr, tagOrBrach string) *GitlabRepoClone {

	localPath := filepath.Join(LOCALPATH_PARATENT_DIR, group, name)
	timeStampStr := time.Now().Local().Format(LayOutSet)

	return &GitlabRepoClone{
		ProjectName:      name,
		Group:            group,
		RepoSSH:          sshAddr,
		TagOrBranch:      tagOrBrach,
		ProjecLocaltPath: localPath,
		TimestampNowDir:  "t_" + timeStampStr,
	}
}

func (g *GitlabRepoClone) Clean() error {

	workDir := filepath.Join(g.ProjecLocaltPath + g.TimestampNowDir)

	if !strings.Contains(workDir, "t_") || !strings.HasPrefix(workDir, LOCALPATH_PARATENT_DIR) {
		return fmt.Errorf("clean dir, check path error: path=%s", workDir)
	}

	cmd := NewCmd("rm", "-rf", workDir)
	return cmd.Run()
}

func (g *GitlabRepoClone) Clone(c ConfigCICD) error {
	workDir := filepath.Join(g.ProjecLocaltPath, g.TimestampNowDir)

	err := os.MkdirAll(workDir, 0755)
	if err != nil {
		slog.Error("os mkdir all", "path", workDir, "msg", err)
		return err
	}

	//clone 之用先清理过多的目录
	fss, err := os.ReadDir(g.ProjecLocaltPath)
	if err != nil {
		return err
	}

	if len(fss) > c.BuildHistoryReserve &&
		strings.Contains(workDir, "t_") &&
		strings.HasPrefix(workDir, LOCALPATH_PARATENT_DIR) {
		//do clean
		for _, f := range fss[:len(fss)-c.BuildHistoryReserve] {
			err := os.RemoveAll(filepath.Join(g.ProjecLocaltPath, f.Name()))
			if err != nil {
				slog.Error("remove dir", "name", f.Name(), "msg", err)
				continue
			}
			slog.Info("remove dir", "path", f.Name(), "msg", "success")
		}
	}

	cloneApp := NewCmd(
		"git", "clone",
		"-b", g.TagOrBranch,
		"--depth=1",
		g.RepoSSH, workDir)

	cloneCICD := NewCmd(
		"git", "clone",
		"-b", "main",
		"--depth=1",
		GITLAB_CICD_REPO_ADDR, filepath.Join(workDir, CICD_REPO_LOCAL_PATH),
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

// cmd with os.Strerr and os.Stdout
func NewCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	// default stderr stdout to /dev/null
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd

}
