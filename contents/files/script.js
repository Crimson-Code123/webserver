
function sendAlert() {
	alert("Works");
}

function getTextElement() {
	document.getElementById("text").innerHTML = " ";
	alert(elmnt.textContent);
}

function print() {
	encodedData = window.btoa("Hello, world");
	console.log(encodedData);
	console.log(window.atob(encodedData));
}
