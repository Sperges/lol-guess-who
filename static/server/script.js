const gameStates = {
	LOBBY: 0,
	PLAYING: 1,
	POST_GAME: 2,
};

let conn;
let champs;
let boardState = new Array(25).fill(false);
let gameState = gameStates.LOBBY;
let clientId;

function cardClicked(index) {
		
}

window.onload = () => {
	const board = document.getElementById("player-board");
	const cards = board.getElementsByClassName("card");
	const cardLabels = board.getElementsByClassName("card-label");
	const otherBoard = document.getElementById("other-board");
	const otherCards = otherBoard.getElementsByClassName("other-card");
	let selectedCard = document.getElementById("selected-card");
	const selectedCardLabel = document.getElementById("selected-card-label");
	const log = document.getElementById("log");
	const input = document.getElementById("msg");

	document.getElementById("form").onsubmit = (evt) => {
		evt.preventDefault();
		chat(input.value);
		input.value = "";
	}

	function appendLog(item) {
        var doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }

	function appendText(text, color="white") {
		var item = document.createElement("div");
		item.innerText = text;
		item.style["color"] = color;
		appendLog(item);
	}

	function updateCardDown(card, down) {
		if (down) {
			card.style["filter"] = "grayscale(100%) blur(5px)";
		} else {
			card.style["filter"] = "grayscale(0%) blur(0px)";
		}
	}

	function populateCards(cardList) {
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
		if (id) {
			selectedCard.style["backgroundImage"] = `url(../images/${id}.jpg)`;
			selectedCardLabel.innerText = NAMES[id];

			selectedCard.addEventListener("click", (evt) => {
				if (confirm("Are you sure you want to reveal your character?")) {
					reveal();
				}
			});
		} else {
			var newSelectedCard = selectedCard.cloneNode(true);
			selectedCard.parentNode.replaceChild(newSelectedCard, selectedCard);
			selectedCard = newSelectedCard;
			selectedCard.style["backgroundImage"] = `url(../images/missing.jpg)`;
			selectedCardLabel.innerText = "Pending";
		}
	}

	function handleMessage(msg) {
		console.log(msg);
		if (msg.boardReset) {
			for (let i = 0; i < boardState.length; i++) {
				gameState = gameStates.LOBBY;
				boardState[i] = false;
				updateCardDown(cards[i], false);
				updateCardDown(otherCards[i], false);
			}
			log.innerHTML = "";
			updateSelectedCard(null);
		}
		if (msg.initialMessage) {
			clientId = msg.initialMessage.clientId;
			champs = msg.initialMessage.champs;
			populateCards(champs);
		}
		if (msg.gameStarted) {
			gameState = gameStates.PLAYING;
			appendText("Starting game!", "green");
		}
		if (msg.flip) {
			const card = otherBoard.getElementsByClassName("other-card")[msg.flip.index];
			updateCardDown(card, msg.flip.down);
		}
		if (msg.reveal) {
			appendText(`Your opponent reveals that their character was ${NAMES[msg.reveal.index]}.`, "yellow");
		}
		if (msg.champSelected) {
			appendText("Your opponent has selected a champ", "yellow");
		}
		if (msg.chat) {
			if (clientId == msg.chat.sender) {
				appendText(`"${msg.chat.text}"`, "gray");
			} else {
				appendText(`"${msg.chat.text}"`);
			}
			
		}
		if (msg.serverFull) {
			appendText(`Two players are connected`, "green");
		}
	}

    if (window["WebSocket"]) {
		const gameid = window.location.href.substring(window.location.href.lastIndexOf("/")+1);
        conn = new WebSocket("ws://" + document.location.host + "/ws/" + gameid);
        conn.onclose = (evt) => {
			console.log("Websocket connection closed");
        };
        conn.onmessage = (evt) => {	
			let messages = 	evt.data.split('\n');
			for (const msg of messages) {
				handleMessage(JSON.parse(msg));
			}
        };
    } else {
		console.log("Your browser does not support WebSockets.");
		appendText("Your browser does not support WebSockets.", "red");
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

const chat = (text) => {
	return sendMessage(JSON.stringify(
		{
			"chat": {
				"text": text,
			}
		}
	));
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

const requestBoardUpdate = () => {
	return sendMessage(JSON.stringify(
		{
			"requestBoardUpdate": {},
		}
	))
}

const requestBoardReset = () => {
	return sendMessage(JSON.stringify(
		{
			"requestBoardReset": {},
		}
	));
}

