package playwright

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/playwright-community/playwright-go"
)

// func assertErrorToNilf(message string, err error) {
// 	if err != nil {
// 		log.Fatalf(message, err)
// 	}
// }

func assertEqual(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		panic(fmt.Sprintf("%v does not equal %v", actual, expected))
	}
}

func isCounterLoaded(page playwright.Page) bool {
	numberOnPageBeforeClickStringIsLoaded, err := page.GetByTestId("countNumber").IsVisible()
	if err != nil {
		fmt.Printf("Can't locate number: %v", err)
	}
	return numberOnPageBeforeClickStringIsLoaded
}

func actualCounterNumber(page playwright.Page) int {
	var numberOnPage int
	if isCounterLoaded(page) {
		numberOnPageBeforeClickString, err := page.GetByTestId("countNumber").TextContent()
		if err != nil {
			fmt.Printf("Can't locate number: %v", err)
		}

		numberOnPage, err = strconv.Atoi(numberOnPageBeforeClickString)
		if err != nil {
			fmt.Printf("Can't convert string: %v", err)
		}
	}
	return numberOnPage
}

func AssertNumbersWhenAmountChangedButtonClicked(page playwright.Page, buttonId string) {
	actualCounterNumberOnPage := actualCounterNumber(page)
	fmt.Println("Before:", actualCounterNumberOnPage)

	page.GetByTestId(buttonId).Click()

	changedNumberOnPage := actualCounterNumber(page)
	fmt.Println("After:", changedNumberOnPage)

	assertEqual(actualCounterNumberOnPage, changedNumberOnPage)

}

func RunTests(test func(page playwright.Page, buttonId string), buttonId string) {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install playwright dependencies: %v", err)
	}
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(
		playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false),
		},
	)
	if err != nil {
		log.Fatalf("Could not launch the browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Could not open a new page: %v", err)
	}

	page.OnRequest(func(request playwright.Request) {
		log.Printf("<< %v %s\n", request.Method(), request.URL())
	})

	page.OnResponse(func(response playwright.Response) {
		log.Printf("<< %v %s\n", response.Status(), response.URL())
	})

	if _, err = page.Goto("http://localhost:3000", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("Could not visit the desired page: %v", err)
	}

	test(page, buttonId)

	if err = browser.Close(); err != nil {
		log.Fatalf("Could not close the desired page: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("Could not stop playwright: %v", err)
	}
}
