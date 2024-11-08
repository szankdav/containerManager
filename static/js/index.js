"use strict";
const baseURL = "http://localhost:8080"

document.getElementById("form").addEventListener("submit", (e) => {
    e.preventDefault();
    repositoryUrl = document.getElementById("repositoryUrl").value
    sendUrlForGo(repositoryUrl);
})

async function sendUrlForGo(url){
    try {
        const response = await fetch(`${baseURL}/url`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
            body: JSON.stringify(url)
        })
        if (!response.ok){
            throw new Error(`Response status: ${response.status}`);
        }
        else{
            return response
        }
    } catch (error) {
        console.error("Error: ", error.message)
    }
}