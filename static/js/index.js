"use strict";
const baseURL = "http://localhost:8080";

const testCounters = new Map([
  ["increaseButton", 0],
  ["decreaseButton", 0],
]);

document.getElementById("form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const repositoryUrl = document.getElementById("repositoryUrl").value;
  const response = await sendUrlForGo(repositoryUrl);
  if (response.status == 201) {
    document.getElementById("repo").style.display = "unset";
    document.getElementById("tests").style.display = "unset";
    response
      .json()
      .then((data) =>
        document.getElementById("visitContainer").setAttribute("href", data)
      );
  } else {
    document.getElementById("error").style.display = "unset";
  }
});

async function runTest(e) {
  const testButtonId = e.target.id;
  console.log(testButtonId);
  try {
    const response = await fetch(`${baseURL}/test`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ buttonID: testButtonId }),
    });
    if (!response.ok) {
      throw new Error(`Response status: ${response.status}`);
    } else {
      response.json().then((data) => {
        if (data == testButtonId) {
          testCounters.set(`${data}`, testCounters.get(`${data}`) + 1);
          document.getElementById(`${data}TestPassed`).innerText =
            testCounters.get(`${data}`);
        }
      });
    }
  } catch (error) {
    console.error("Error: ", error.message);
  }
}

const testButtons = document.getElementsByClassName("testButton");
for (let i = 0; i < testButtons.length; i++) {
  testButtons[i].addEventListener("click", runTest);
}

async function sendUrlForGo(url) {
  try {
    const response = await fetch(`${baseURL}/url`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({ bodyURL: url }),
    });
    if (!response.ok) {
      throw new Error(`Response status: ${response.status}`);
    } else {
      return response;
    }
  } catch (error) {
    console.error("Error: ", error.message);
  }
}
