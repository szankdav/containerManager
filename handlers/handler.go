package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

func ConnectDocker(c *gin.Context) *client.Client {
	//Connect to client (docker engine)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return cli
}

func BuildDockerImage(c *gin.Context) {
	cli := ConnectDocker(c)
	ctx := context.Background()

	images, err := cli.ImageList(ctx, image.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		fmt.Println(image.ID)
	}
	defer cli.Close()
}

func CloneRepositoryWithUrl(url string, repoDirectory string) {
	cmdStruct := exec.Command("powershell.exe", "git clone", url)
	out, err := cmdStruct.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(out))
	cmdStruct.Dir = repoDirectory
	fmt.Print(repoDirectory)
}

func getRepoDirectory(url string) string {
	repoNameSlice := strings.Split(url, "/")
	repoDirectoryNameWithGit := strings.Split(repoNameSlice[len(repoNameSlice)-1], ".")
	repoDirectoryNameWithoutGit := repoDirectoryNameWithGit[0]
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return workingDirectory + "\\" + repoDirectoryNameWithoutGit
}

func GetUrlFromHeader(c *gin.Context) {
	reqBody, err := io.ReadAll(c.Request.Body)
	urlFromBody := ""
	if err != nil {
		fmt.Println("Error reading body:", err)
	}
	json.Unmarshal(reqBody, &urlFromBody)
	c.Request.Body.Close()
	CloneRepositoryWithUrl(urlFromBody, getRepoDirectory(urlFromBody))
}
