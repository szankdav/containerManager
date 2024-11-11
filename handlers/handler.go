package handler

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
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

func BuildDockerImage(c *gin.Context, tags []string, dockerfile string) error {
	cli := ConnectDocker(c)
	ctx := context.Background()

	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	// Create a filereader
	dockerFileReader, err := os.Open(dockerfile)
	if err != nil {
		return err
	}

	// Read the actual Dockerfile
	readDockerFile, err := io.ReadAll(dockerFileReader)
	if err != nil {
		return err
	}

	// Make a TAR header for the file
	tarHeader := &tar.Header{
		Name: dockerfile,
		Size: int64(len(readDockerFile)),
	}

	// Writes the header described for the TAR file
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}

	// Writes the dockerfile data to the TAR file
	_, err = tw.Write(readDockerFile)
	if err != nil {
		return err
	}

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	// Define the build options to use for the file
	buildOptions := types.ImageBuildOptions{
		Context:    dockerFileTarReader,
		Dockerfile: dockerfile,
		Remove:     true,
		Tags:       tags,
	}

	// Build the actual image
	imageBuildResponse, err := cli.ImageBuild(
		ctx,
		dockerFileTarReader,
		buildOptions,
	)

	if err != nil {
		return err
	}

	// Read the STDOUT from the build process
	defer imageBuildResponse.Body.Close()
	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		return err
	}

	defer cli.Close()
	return nil

}

func CloneRepositoryWithUrl(url string) {
	cmdStruct := exec.Command("git", "clone", url)
	out, err := cmdStruct.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(out))
}

func GetRepoFolderName(url string) string {
	repoNameSlice := strings.Split(url, "/")
	repoDirectoryNameWithGit := strings.Split(repoNameSlice[len(repoNameSlice)-1], ".")
	repoDirectoryNameWithoutGit := repoDirectoryNameWithGit[0]

	return repoDirectoryNameWithoutGit
}

func GetUrlFromHeader(c *gin.Context) (string, error) {
	reqBody, err := io.ReadAll(c.Request.Body)
	urlFromBody := ""
	if err != nil {
		fmt.Println("Error reading body:", err)
	}
	json.Unmarshal(reqBody, &urlFromBody)
	c.Request.Body.Close()
	return urlFromBody, err
}

func StartContainer(c *gin.Context) {
	url, err := GetUrlFromHeader(c)
	if err != nil {
		fmt.Println(err)
	}
	CloneRepositoryWithUrl(url)
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	repoFolderName := GetRepoFolderName(url)
	tags := []string{strings.ToLower(repoFolderName)}
	dockerfile := workingDirectory + "/" + repoFolderName + "/Dockerfile"
	err = BuildDockerImage(c, tags, dockerfile)
	if err != nil {
		log.Println(err)
	}

	cmdStruct := exec.Command("docker", "images")
	out, err := cmdStruct.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(out))
}
