const testChamps = [99,31,73,149,106,51,40,61,3,108,157,48,64,160,140,123,17,152,109,113,18,95,45,37,161];

window.onload = () => {
	function createCard(id) {
		return `
		<div class="card">
			<label class="label">${NAMES[id]}</label>
			<img class="image" id="${NAMES[id].toLowerCase()}" src="../images/${id}.jpg" alt="${NAMES[id]}"/>
		</div>
		`
	}

	function updateSelectedCard(id) {
		selectedCard.style["backgroundColor"] = null;
		selectedCardImage.src = `../images/${id}.jpg`;
		selectedCardImage.alt = NAMES[id];
		selectedCardLabel.innerText = NAMES[id];
	}

	const board = document.getElementById("player-board");
	const selectedCard = document.getElementById("selected-card");
	const selectedCardImage = document.getElementById("selected-card-img");
	const selectedCardLabel = document.getElementById("selected-card-label");

	// updateSelectedCard(testChamps[3]);
	

	console.log("populating cards");

	// testChamps.forEach(id => board.innerHTML += createCard(id))
}