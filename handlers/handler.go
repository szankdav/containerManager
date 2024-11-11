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
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
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

func BuildDockerImage(c *gin.Context, tags []string, dockerFolder string) error {
	cli := ConnectDocker(c)
	ctx := context.Background()
	// Dockerfile path
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

func RunContainer(c *gin.Context, imageName string) {
	cli := ConnectDocker(c)
	ctx := context.Background()
	defer cli.Close()

	// Pull the image
	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		fmt.Println(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	// Set the host
	hostConfig := container.HostConfig{}

	// Create the container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			"3000/tcp": struct{}{},
		},
	}, &hostConfig, nil, nil, "")
	if err != nil {
		fmt.Println(err)
	}

	// Start the container
	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		fmt.Println(err)
	}

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
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	// Save the home directory
	workingDirectory := currentWorkingDirectory

	repoFolderName := GetRepoFolderName(url)
	tags := []string{strings.ToLower(repoFolderName)}
	dockerFolder := currentWorkingDirectory + "/" + repoFolderName

	// Change the current working directory to the repos path
	os.Chdir(dockerFolder)
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mydir)

	// Build the image
	err = BuildDockerImage(c, tags, mydir)
	if err != nil {
		log.Println(err)
	}

	// List the images
	cmdStruct := exec.Command("docker", "images")
	out, err := cmdStruct.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(out))

	// Change back to home directory
	defer os.Chdir(workingDirectory)

	defer RunContainer(c, tags[0])

}
