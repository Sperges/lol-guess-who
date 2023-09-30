const testChamps = [99,31,73,149,106,51,40,61,3,108,157,48,64,160,140,123,17,152,109,113,18,95,45,37,161];

window.onload = () => {
	function populateCards(cardList) {
		const board = document.getElementById("player-board");
		const cards = board.getElementsByClassName("card");
		const cardLabels = board.getElementsByClassName("card-label");
		for (let i = 0; i < cards.length; i++) {
			const id = cardList[i];
			cards[i].style["backgroundImage"] = `url(../images/${id}.jpg)`;
			cardLabels[i].innerText = NAMES[id];
		}
	} 

	function updateSelectedCard(id) {
		const selectedCard = document.getElementById("selected-card");
		const selectedCardLabel = document.getElementById("selected-card-label");
		selectedCard.style["backgroundImage"] = `url(../images/${id}.jpg)`;
		selectedCardLabel.innerText = NAMES[id];
	}

	console.log("populating cards");

	populateCards(testChamps);

	updateSelectedCard(testChamps[5]);
}