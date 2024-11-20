package handler

import (
	util "example.com/container-manager/internal/util"
	playwright "example.com/container-manager/playwright"

	"fmt"

	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// This is the actual entrypoint
func StartContainer(c *gin.Context) {
	// Get current working directory
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	// Get the URL for the repo we want to clone
	url, err := util.GetUrlFromBody(c)
	if err != nil {
		fmt.Println(err)
	}

	// Clone the repo with the URL
	util.CloneRepositoryWithUrl(url)
	repoFolderName := util.GetRepoFolderName(url)
	tags := []string{strings.ToLower(repoFolderName)}

	// Change working directory to the cloned repos folder
	util.ChangeWorkingDirectory(currentWorkingDirectory, repoFolderName)

	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}

	// If docker image is already built, just run the container
	if util.CheckIfDockerImageAlreadyExists(tags) {
		util.RunDockerImage(tags)
	} else {
		// Build the image
		err = util.BuildDockerImage(c, tags, mydir)
		if err != nil {
			log.Println(err)
		}
		// Run the container
		util.RunDockerImage(tags)
	}

	// Change back to home directory
	defer os.Chdir(currentWorkingDirectory)

	containerURL := "http://localhost:3000"

	//Give back the URL where we can reach the container
	c.JSON(http.StatusCreated, containerURL)
}

func SpinUpTest(c *gin.Context) {
	buttonId, err := util.GetButtonIdFromBody(c)
	if err != nil {
		fmt.Println(err)
	}
	playwright.RunTests(playwright.AssertNumbersWhenAmountChangedButtonClicked, buttonId)
	c.JSON(http.StatusOK, buttonId)
}
