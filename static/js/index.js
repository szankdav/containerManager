"use strict";
const baseURL = "http://localhost:8080";

let increaseTestPassed = 0;
let decreaseTestPassed = 0;

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

document.getElementById("increaseTest").addEventListener("click", async () => {
  try {
    const response = await fetch(`${baseURL}/increaseTest`)
    if(!response.ok){
      throw new Error(`Response status: ${response.status}`);
    } else {
      increaseTestPassed += 1;
      document.getElementById("increaseTestPassed").innerText = increaseTestPassed
      return response;
    }
  } catch (error) {
    console.error("Error: ", error.message);
  }
})

document.getElementById("decreaseTest").addEventListener("click", async () => {
  try {
    const response = await fetch(`${baseURL}/decreaseTest`)
    if(!response.ok){
      throw new Error(`Response status: ${response.status}`);
    } else {
      decreaseTestPassed += 1;
      document.getElementById("decreaseTestPassed").innerText = decreaseTestPassed
      return response;
    }
  } catch (error) {
    console.error("Error: ", error.message);
  }
})

async function sendUrlForGo(url) {
  try {
    const response = await fetch(`${baseURL}/url`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify(url),
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
