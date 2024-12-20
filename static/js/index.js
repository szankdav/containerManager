"use strict";
const baseURL = "http://localhost:8080";

document.getElementById("form").addEventListener("submit", async (e) => {
  e.preventDefault();
  const repositoryUrl = document.getElementById("repositoryUrl").value;
  const response = await sendUrlForGo(repositoryUrl);
  if (response.status == 201) {
    document.getElementById("repo").style.display = "unset";
    response
      .json()
      .then((data) =>
        document.getElementById("visitContainer").setAttribute("href", data)
      );
  } else {
    document.getElementById("error").style.display = "unset";
  }
});

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
