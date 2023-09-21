
let conn;

window.onload = function () {
    let msg = document.getElementById("msg");
    let log = document.getElementById("log");
	let rid = window.location.href.substring(window.location.href.lastIndexOf("/")+1)

    appendLog = (item) => {
        let doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }

    document.getElementById("form").onsubmit = () => {
        if (!conn) {
            return false;
        }
        if (!msg.value) {
            return false;
        }
        conn.send(msg.value);
        msg.value = "";
        return false;
    };

    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws/" + rid);

        conn.onclose = (evt) => {
            let item = document.createElement("div");
            item.innerHTML = "<b>Connection closed.</b>";
            appendLog(item);
        };

        conn.onmessage = (evt) => {
            let messages = evt.data.split('\n');
            for (let i = 0; i < messages.length; i++) {
                let item = document.createElement("div");
                item.innerText = messages[i];
                appendLog(item);
            }
        };

    } else {
        let item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
    }

	flipDown = (index) => {
		flip(index, false);
	}

	flipUp = (index) => {
		flip(index, true);
	}

	flip = (index, up) => {
		let msg = getFlipMessage(index, up);
		if (!conn) {
			return false;
		}
		conn.send(msg);
		return true;
	}

	reveal = (index) => {
		let msg = getRevealMessage(index);
		if (!conn) {
			return false;
		}
		conn.send(msg);
		return true;
	}
	
	getFlipMessage = (index, up) => {
		return JSON.stringify(
			{
				"type": "flip",
				"index": index,
				"state": up,
			}
		);
	}

	getSelectChampMessage = (index) => {
		return JSON.stringify(
			{
				"type": "select",
				"index": index,
			}
		);
	}
	
	getRevealMessage = (index) => {
		return JSON.stringify(
			{
				"type": "reveal",
				"index": index,
			}
		);
	}
};

