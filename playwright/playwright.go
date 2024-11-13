package playwright

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/playwright-community/playwright-go"
)

func assertErrorToNilf(message string, err error) {
	if err != nil {
		log.Fatalf(message, err)
	}
}

func assertEqual(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		panic(fmt.Sprintf("%v does not equal %v", actual, expected))
	}
}

func assertNumbersWhenIncreaseClicked(page playwright.Page, btnOneId string, numberFieldId string) {
	numberOnPageBeforeClickString, err := page.Locator("." + numberFieldId).TextContent()
	if err != nil {
		fmt.Printf("Can't locate number: %v", err)
	}
	numberOnPageBeforeClickInt, err := strconv.Atoi(numberOnPageBeforeClickString)
	if err != nil {
		fmt.Printf("Can't convert string: %v", err)
	}
	buttonOne := page.Locator("." + btnOneId)
	buttonOne.Click()
	numberOnPageAfterClickString, err := page.Locator("." + numberFieldId).TextContent()
	if err != nil {
		fmt.Printf("Can't locate number: %v", err)
	}
	numberOnPageAfterClickInt, err := strconv.Atoi(numberOnPageAfterClickString)
	if err != nil {
		fmt.Printf("Can't convert string: %v", err)
	}
	assertErrorToNilf("could not determine the count amount: %w", err)
	assertEqual(numberOnPageBeforeClickInt, numberOnPageAfterClickInt)
}

func ButtonClickTest(url string, btnOneId string, btnTwoId string, numberFieldId string) {
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

	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("Could not visit the desired page: %v", err)
	}

	assertNumbersWhenIncreaseClicked(page, btnOneId, numberFieldId)

	if err = browser.Close(); err != nil {
		log.Fatalf("Could not close the desired page: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("Could not stop playwright: %v", err)
	}
}
