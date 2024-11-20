package util

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
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func GetUrlFromBody(c *gin.Context) (string, error) {
	reqBody, err := io.ReadAll(c.Request.Body)
	var body map[string]string
	if err != nil {
		fmt.Println("Error reading body:", err)
	}
	json.Unmarshal(reqBody, &body)
	c.Request.Body.Close()
	return body["bodyURL"], err
}

func GetButtonIdFromBody(c *gin.Context) (string, error) {
	reqBody, err := io.ReadAll(c.Request.Body)
	var body map[string]string
	if err != nil {
		fmt.Println("Error reading body:", err)
	}
	json.Unmarshal(reqBody, &body)
	c.Request.Body.Close()
	return body["buttonID"], err
}

func CheckIfRepoAlreadyCloned(repoFolderName string) bool {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	if _, err := os.Stat(currentWorkingDirectory + "/" + repoFolderName); !os.IsNotExist(err) {
		return true
	}
	return false
}

func CloneRepositoryWithUrl(url string) {
	if !CheckIfRepoAlreadyCloned(GetRepoFolderName(url)) {
		cmdStruct := exec.Command("git", "clone", url)
		out, err := cmdStruct.Output()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Print(string(out))
	}
}

func GetRepoFolderName(url string) string {
	repoNameSlice := strings.Split(url, "/")
	repoDirectoryNameWithGit := strings.Split(repoNameSlice[len(repoNameSlice)-1], ".")
	repoDirectoryNameWithoutGit := repoDirectoryNameWithGit[0]

	return repoDirectoryNameWithoutGit
}

func ConnectDocker(c *gin.Context) *client.Client {
	//Connect to client (docker engine)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return cli
}

// Build image from the pulled repo
func BuildDockerImage(c *gin.Context, tags []string, dockerFolder string) error {
	cli := ConnectDocker(c)
	ctx := context.Background()

	// Dockerfile path in the pulled folder
	dockerfile := dockerFolder + "/Dockerfile"

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

	// Add every file around the Dockerfile to the .tar file
	tw.AddFS(os.DirFS(dockerFolder))

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

func RunDockerImage(tags []string) {
	// Start container with testcontainers
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: tags[0],
			ConfigModifier: func(config *container.Config) {
				config.Env = []string{"a=b"}
			},
			HostConfigModifier: func(hostConfig *container.HostConfig) {
				hostConfig.PortBindings = nat.PortMap{
					"3000/tcp": []nat.PortBinding{
						{
							HostIP:   "0.0.0.0",
							HostPort: "3000",
						},
					},
				}
			},
			ExposedPorts: []string{"3000/tcp"},
			WaitingFor:   wait.ForListeningPort("3000/tcp"),
		},
		Started: true,
	})
	if err != nil {
		log.Println(err)
	}

	fmt.Println(container.ContainerIP(ctx))
}

func CheckIfDockerImageAlreadyExists(tags []string) bool {
	// List the images, so we can make sure the image is created
	var imageExist bool = false
	cmdStruct := exec.Command("docker", "images")
	out, err := cmdStruct.Output()
	if err != nil {
		fmt.Println(err)
	}

	if strings.Contains(string(out), tags[0]) {
		imageExist = true
	}

	fmt.Print(string(out))

	return imageExist
}

func ChangeWorkingDirectory(currentWorkingDirectory string, repoFolderName string) {
	// Save the name of the cloned repo folder
	var dockerFolder string = currentWorkingDirectory + "/" + repoFolderName

	// Change the current working directory to the repo folder
	os.Chdir(dockerFolder)
}
