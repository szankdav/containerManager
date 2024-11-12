package handler

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/docker/docker/api/types"
	network "github.com/docker/docker/api/types/network"
	natting "github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
)

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

func RunContainer(c *gin.Context, imageName string, port string, inputEnv []string) error {
	cli := ConnectDocker(c)
	ctx := context.Background()
	defer cli.Close()

	// Define a PORT opening
	newport, err := natting.NewPort("tcp", port)
	if err != nil {
		fmt.Println("Unable to create docker port")
		return err
	}

	// Configured hostConfig:
	hostConfig := &container.HostConfig{
		PortBindings: natting.PortMap{
			newport: []natting.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	// Define Network config:
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	gatewayConfig := &network.EndpointSettings{
		Gateway: "gatewayname",
	}
	networkConfig.EndpointsConfig["bridge"] = gatewayConfig

	// Define ports to be exposed (has to be same as hostconfig.portbindings.newport)
	exposedPorts := map[natting.Port]struct{}{
		newport: struct{}{},
	}

	// Configuration for the container create
	config := &container.Config{
		Image:        imageName,
		Env:          inputEnv,
		ExposedPorts: exposedPorts,
		Hostname:     fmt.Sprintf("%s-hostnameexample", imageName),
	}

	// Creating the actual container
	cont, err := cli.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkConfig,
		nil,
		imageName,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	// Run the actual container
	cli.ContainerStart(ctx, cont.ID, container.StartOptions{})
	log.Printf("Container %s is created", cont.ID)

	return nil
}

// This is the actual entrypoint
func StartContainer(c *gin.Context) {
	// Get the URL for the repo we want to clone
	url, err := GetUrlFromHeader(c)
	if err != nil {
		fmt.Println(err)
	}

	// Clone the repo with the URL
	CloneRepositoryWithUrl(url)
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	// Save the home directory for later
	workingDirectory := currentWorkingDirectory

	// Save the name of the cloned repo folder
	repoFolderName := GetRepoFolderName(url)
	tags := []string{strings.ToLower(repoFolderName)}
	dockerFolder := currentWorkingDirectory + "/" + repoFolderName

	// Change the current working directory to the repo folder
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

	// List the images, so we can make sure the image is created
	cmdStruct := exec.Command("docker", "images")
	out, err := cmdStruct.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Print(string(out))

	// Change back to home directory
	defer os.Chdir(workingDirectory)

	// Define the port we want to use on local
	portopening := "3000"
	inputEnv := []string{fmt.Sprintf("LISTENINGPORT=%s", portopening)}

	// Run the container
	err = RunContainer(c, tags[0], portopening, inputEnv)

	if err != nil {
		log.Println(err)
	}

	containerURL := "http://localhost:" + portopening

	c.JSON(http.StatusCreated, containerURL)
}
