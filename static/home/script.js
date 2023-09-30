const gameStates = {
	LOBBY: 0,
	PLAYING: 1,
	POST_GAME: 2,
};

let conn;
let champs;
let boardState = new Array(25).fill(false);
let playerId;
let gameState = gameStates.LOBBY;

function cardClicked(index) {
		
}

window.onload = () => {

	const otherBoard = document.getElementById("other-board");

	function updateCardDown(card, down) {
		if (down) {
			card.style["filter"] = "grayscale(100%) blur(5px)";
		} else {
			card.style["filter"] = "grayscale(0%) blur(0px)";
		}
	}

	function populateCards(cardList) {
		const board = document.getElementById("player-board");
		const cards = board.getElementsByClassName("card");
		const cardLabels = board.getElementsByClassName("card-label");

		for (let i = 0; i < cards.length; i++) {
			const card = cards[i];
			const id = cardList[i];
			card.style["backgroundImage"] = `url(../images/${id}.jpg)`;
			cardLabels[i].innerText = NAMES[id];

			card.addEventListener("click", (evt) => {
				let index = i;
				switch (gameState) {
					case gameStates.LOBBY:
						updateSelectedCard(champs[index]);
						select(id);
						break;
					default:
						boardState[index] = !boardState[index];
						const down = boardState[index]
						updateCardDown(card, down);
						flip(index, down);
						break;
				}
			});
		}
	}

	function updateSelectedCard(id) {
		const selectedCard = document.getElementById("selected-card");
		const selectedCardLabel = document.getElementById("selected-card-label");
		selectedCard.style["backgroundImage"] = `url(../images/${id}.jpg)`;
		selectedCardLabel.innerText = NAMES[id];

		selectedCard.addEventListener("click", (evt) => {
			if (confirm("Are you sure you want to reveal your character?")) {
				reveal();
			}
		});
	}

	function handleMessage(msg) {
		console.log(msg);
		if (msg.initialMessage) {
			playerId = msg.initialMessage.playerid;
			champs = msg.initialMessage.champs;
			populateCards(champs);
		}
		if (msg.gameStarted) {
			gameState = gameStates.PLAYING;
		}
		if (msg.flip) {
			const card = otherBoard.getElementsByClassName("other-card")[msg.flip.index];
			updateCardDown(card, msg.flip.down);
		}
		if (msg.reveal) {
			alert(`Your opponent reveals that their character was ${NAMES[msg.reveal.index]}.`);
		}
	}

    if (window["WebSocket"]) {
		const gameid = window.location.href.substring(window.location.href.lastIndexOf("/")+1);
        conn = new WebSocket("ws://" + document.location.host + "/ws/" + gameid);
        conn.onclose = (evt) => {
			console.log("Websocket connection closed");
        };
        conn.onmessage = (evt) => {		
			handleMessage(JSON.parse(evt.data));
        };
    } else {
		console.log("Your browser does not support WebSockets.");
    }
};

const sendMessage = (msg) => {
	if (!conn) {
		return false;
	}
	if (!msg) {
		return false;
	}
	conn.send(msg);
	return true;
}

const select = (index) => {
	
	return sendMessage(JSON.stringify(
		{
			"requestSelectChamp": {
				"index": index,
			}
		}
	))
}

const flip = (index, down) => {
	return sendMessage(JSON.stringify(
		{
			"flip": {
				"index": index,
				"down": down,
			}
		}
	));
}

const reveal = () => {
	return sendMessage(JSON.stringify(
		{
			"reveal": {},
		}
	))
}

const chat = (msg) => {
	return sendMessage(JSON.stringify(
		{
			"chat": {
				"text": msg,
			}
		}
	))
}

const requestBoardUpdate = () => {
	return sendMessage(JSON.stringify(
		{
			"requestBoardUpdate": {},
		}
	))
}

