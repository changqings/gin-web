package cicd

import (
	"bufio"
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	k8scrdClient "github.com/changqings/k8scrd/client"
	"github.com/gin-gonic/gin"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	LocalDeployBaseDir = "/data/devops/deploy/"
)

type K8sDeploy struct {
	AppName   string                `json:"app_name"`
	Group     string                `json:"group"`
	Namespace string                `josn:"namespace"`
	WorkDir   string                `json:"-"`
	Type      string                `json:"project_type"`
	Env       string                `json:"project_env"`
	Tag       string                `json:"tag_or_branch"`
	Image     string                `json:"-"`
	Client    *kubernetes.Clientset `json:"-"`
}

func NewK8sDeploy(name, group, namespace, deployType, env, tag string) *K8sDeploy {
	client := k8scrdClient.GetClient()
	image := filepath.Join(TxTcrHost, TxTcrNamespaceDevopsScq, name+":"+env+"-"+tag)
	timeStampStr := "t_" + time.Now().Local().Format(TimeLayOutSet)
	workDir := filepath.Join(LocalDeployBaseDir, namespace, name, timeStampStr)
	return &K8sDeploy{
		AppName:   name,
		Group:     group,
		Namespace: namespace,
		Type:      deployType,
		Env:       env,
		Tag:       tag,
		Image:     image,
		WorkDir:   workDir,
		Client:    client,
	}

}

func (k *K8sDeploy) SetAll() {
	client := k8scrdClient.GetClient()
	image := filepath.Join(TxTcrHost, TxTcrNamespaceDevopsScq, k.AppName+":"+k.Env+"-"+k.Tag)
	timeStampStr := "t_" + time.Now().Local().Format(TimeLayOutSet)
	workDir := filepath.Join(LocalDeployBaseDir, k.Namespace, k.AppName, timeStampStr)

	k.Client = client
	k.Image = image
	k.WorkDir = workDir

}

// if found on k8s update image
// or depoly use devops/cicd project/deploy/ yaml files, like kustomize/helm/operator
func (k *K8sDeploy) DoDeploy() error {

	deploy, err := k.Client.AppsV1().Deployments(k.Namespace).Get(context.Background(), k.AppName, metav1.GetOptions{ResourceVersion: "0"})
	if err != nil {
		if !k8sErrors.IsNotFound(err) {
			slog.Error("DoDeploy get deploy", "namespace", k.Namespace, "name", k.AppName, "msg", err)
			return err
		}
	}

	// deploy为结构体指针，不能使用nil判断，可以随便取一个结构体字段的值来进行判断是否为空
	if deploy.Name != "" {
		slog.Info("DoDeploy update deployment", "namespace", k.Namespace, "name", k.AppName, "msg", "founded, update image...")
		var appContainerIndex int
		for i, c := range deploy.Spec.Template.Spec.Containers {
			if c.Name == "app" {
				appContainerIndex = i
				break
			}
		}

		deploy.ResourceVersion = ""
		deploy.Spec.Template.Spec.Containers[appContainerIndex].Image = k.Image

		_, err := k.Client.AppsV1().Deployments(k.Namespace).Update(context.Background(), deploy, metav1.UpdateOptions{})
		if err != nil {
			slog.Error("update deploy", "namespace", k.Namespace, "name", k.AppName, "msg", err)
			return err
		}

		slog.Info("DoDeploy update deployment", "namespace", k.Namespace, "name", k.AppName, "msg", "success.")

	} else {
		slog.Info("DoDeploy not found on k8s cluster", "namespace", k.Namespace, "name", k.AppName, "msg", "use devops/cicd config...")
		// mkdir -p k.Workdir
		err := os.MkdirAll(k.WorkDir, 0755)
		if err != nil {
			slog.Error("deploy mkdir all", "path", k.WorkDir, "msg", err)
			return err
		}

		cloneCICDCmd := NewCmd(
			"git", "clone",
			"-b", "main",
			"--depth=1",
			GitlabCICDRepoAddr, k.WorkDir)

		err = cloneCICDCmd.Run()
		if err != nil {
			slog.Error("deploy clone cicd", "path", k.WorkDir, "msg", err)
			return err
		}

		// do deploy
		k8sDeployYamlPath := filepath.Join(k.WorkDir, "jobs", k.Group, k.AppName, "deploy", k.Env)

		//部署使用了模板文件deployment.yaml，占位符预设为#{AppName},#{Namespace},#{Image}
		deploymentYaml := "deployment.yaml"
		preReserveStrs := map[string]string{
			"#{AppName}":   k.AppName,
			"#{Namespace}": k.Namespace,
			"#{Image}":     k.Image}

		deploymentYamlPath := filepath.Join(k8sDeployYamlPath, deploymentYaml)

		for oldStr, newStr := range preReserveStrs {
			err := ReplaceAllWithFile(deploymentYamlPath, oldStr, newStr)
			if err != nil {
				slog.Error("DoDeploy replace all pre-reserve str", "msg", err)
				return err
			}
		}

		deployCmd := NewCmd(
			"/bin/sh", "-c", "kubectl apply -f"+" "+deploymentYaml,
		)
		deployCmd.Dir = k8sDeployYamlPath
		if err := deployCmd.Run(); err != nil {
			slog.Error("deploy run", "cmd", deployCmd.String(), "msg", err)
			return err
		}

		slog.Info("DoDeploy create deployment", "namespace", k.Namespace, "name", k.AppName, "msg", "success.")
	}

	return nil
}

// gin handlefunc
func (k *K8sDeploy) Deploy() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		// bind
		if err := ctx.ShouldBind(&k); err != nil {
			slog.Error("deploy bind to &K8sDeploy{}", "msg", err)
			ctx.AbortWithError(482, errors.New("post body json not right for deploy, please check"))
			return
		}

		// setall
		k.SetAll()

		// deploy
		errDeploy := k.DoDeploy()
		if errDeploy != nil {
			slog.Error("main k8s deploy", "namespace", k.Namespace, "name", k.AppName,
				"env", k.Env, "msg", errDeploy)
			return
		}

		ctx.JSON(200, gin.H{
			"project_name":  k.AppName,
			"project_group": k.Group,
			"project_env":   k.Env,
			"push_image":    k.Image,
			"msg":           "deploy to k8s success.",
		})

	}

}

// replace func, generate by gpt4.0
func ReplaceAllWithFile(filePath string, oldString, newString string) error {
	// 打开源文件
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "tmp-*")
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	// 创建bufio的Scanner来逐行读取文件内容
	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(tmpFile)

	// 逐行读取并替换
	for scanner.Scan() {
		line := scanner.Text()
		// 执行替换操作
		replacedLine := strings.ReplaceAll(line, oldString, newString)
		// 将替换后的行写入临时文件
		_, err := writer.WriteString(replacedLine + "\n") // 添加换行符
		if err != nil {
			return err
		}
	}

	// 确保所有输出被刷新到临时文件
	if err := writer.Flush(); err != nil {
		return err
	}

	// 关闭源文件和临时文件
	if err := file.Close(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// 替换原文件
	if err := os.Rename(tmpFile.Name(), filePath); err != nil {
		return err
	}

	return nil
}
