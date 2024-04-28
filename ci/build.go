package ci

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

// write this ci_cd just for learn use golang.

// run in docker and only on linux, code on ubuntu
// 过程：代码仓库(gitlab_on_docker) --> 挂载到特定容器(docker local) --> 构建并生成产物(use Dockerfile)
// --> 构建镜像并上传到镜像仓库(docker-registry)

// gitlab_admin_read_all token=glpat-Arnoz3_yrcDhZyVgYt9j
// gitlab_host = http://192.168.1.15

// ssh://git@gitlab.scq.com:522/devops/go-ws.git
const (
	GITLAB_HTTP_ADDR            = "gitlab.scq.com"
	GITLAB_SSH_ADDR             = "gitlab.scq.com:522"
	TX_TCR_HOST                 = "ccr.ccs.tencentyun.com"
	TX_TCR_NAMESPACE_DEVOPS_SCQ = "devops_scq"
	LOCALPATH_PARATENT_DIR      = "/data/build/gitlab/"
)

var (
	LayOutSet = "20060102150405.000"
)

// use ~/.ssh/id_rsa as git repo private key
type GitlabRepoClone struct {
	ProjectName string
	Group       string
	RepoSSH     string
	TagOrBranch string
	LocalPath   string
}

type DockerBuild struct {
	ProjectName      string
	ProjectLocalPath string
	DockerfileName   string
	buildType        string
	BuildEnv         string
	BuildTAg         string
}

type DockerPush struct {
	ProjectName     string
	RegistryAddress string
	RegistryAuth    string
}

func (d *DockerBuild) Do() error {

	cmd := NewCmd("docker", "build",
		"-t", TX_TCR_HOST+"/"+TX_TCR_NAMESPACE_DEVOPS_SCQ+"/"+d.ProjectName+":"+d.BuildTAg,
		"-f", d.DockerfileName,
		".",
	)
	cmd.Dir = d.ProjectLocalPath

	return cmd.Run()

}

func NewDockerBuild(name, path, dockerfileName, buildType, buildTag, buildEnv string) *DockerBuild {
	return &DockerBuild{
		ProjectName:      name,
		ProjectLocalPath: path,
		DockerfileName:   dockerfileName,
		buildType:        buildType,
		BuildEnv:         buildEnv,
		BuildTAg:         buildTag,
	}
}

func NewGitlabRepoClone(group, name, sshAddr, tagOrBrach string) *GitlabRepoClone {

	localPath := LOCALPATH_PARATENT_DIR + group + "/" + name + "/"
	timeStampStr := time.Now().Local().Format(LayOutSet)

	return &GitlabRepoClone{
		ProjectName: name,
		Group:       group,
		RepoSSH:     sshAddr,
		TagOrBranch: tagOrBrach,
		LocalPath:   localPath + "t_" + timeStampStr,
	}
}

func (g *GitlabRepoClone) Clean() error {

	if !strings.Contains(g.LocalPath, "t_") && strings.HasPrefix(g.LocalPath, LOCALPATH_PARATENT_DIR) {
		return fmt.Errorf("clean dir, check path error: path=%s", g.LocalPath)
	}

	cmd := NewCmd("rm", "-rf", g.LocalPath)
	return cmd.Run()
}

func (g *GitlabRepoClone) Do() error {

	err := os.MkdirAll(g.LocalPath, 0755)
	if err != nil {
		slog.Error("os mkdir all", "path", g.LocalPath, "msg", err)
		return err
	}

	cmd := NewCmd("git", "clone",
		"-b", g.TagOrBranch,
		"--depth=1",
		g.RepoSSH, g.LocalPath)

	// debug cmd
	fmt.Printf("cmd.String()=%s\n", cmd.String())

	return cmd.Run()
}

// cmd with os.Strerr and os.Stdout
func NewCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)

	// default stderr stdout to /dev/null
	cmd.Stderr = os.Stderr
	cmd.Stderr = os.Stdout

	return cmd

}
