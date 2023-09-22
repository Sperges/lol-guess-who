let conn;

window.onload = function () {
    const msg = document.getElementById("msg");
    const log = document.getElementById("log");
	const gameid = window.location.href.substring(window.location.href.lastIndexOf("/")+1);

    const appendLog = (item) => {
        const doScroll = log.scrollTop > log.scrollHeight - log.clientHeight - 1;
        log.appendChild(item);
        if (doScroll) {
            log.scrollTop = log.scrollHeight - log.clientHeight;
        }
    }

    document.getElementById("form").onsubmit = (evt) => {
		evt.preventDefault();
        const result = sendMessage(msg.value);
        msg.value = "";
        return result;
    };

    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws/" + gameid);
        conn.onclose = (evt) => {
            const item = document.createElement("div");
            item.innerHTML = "<b>Connection closed.</b>";
            appendLog(item);
        };
        conn.onmessage = (evt) => {
            const messages = evt.data.split('\n');
            for (let i = 0; i < messages.length; i++) {
                const item = document.createElement("div");
                item.innerText = messages[i];
                appendLog(item);
            }
        };
    } else {
        const item = document.createElement("div");
        item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
        appendLog(item);
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

const select = (id) => {
	return sendMessage(JSON.stringify(
		{
			"type": "select",
			"selection": id,
		}
	))
}

const flip = (index, up) => {
	return sendMessage(JSON.stringify(
		{
			"type": "flip",
			"index": index,
			"up": up,
		}
	));
}

const reveal = () => {
	return sendMessage(JSON.stringify(
		{
			"type": "reveal",
		}
	))
}