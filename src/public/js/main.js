let $scoreBtn = $('.btn-scope');
let ws;

$scoreBtn.click((e) => {
    let score = e.currentTarget.getAttribute('data');
    ws.send(JSON.stringify({score: score}));
});

function ws_connect() {
    const params = window.location.href.split("/");
    const roomId = params[params.length - 1];

    ws = new WebSocket("ws://" + document.location.host + "/ws/" + roomId);

    ws.onopen = () => {
        console.log("Connection open ...");
        /*let obj = {score: -1};
        ws.send(JSON.stringify(obj));*/
    };

    ws.onmessage = (evt) => {
        console.log("Received Message: " + evt.data);
        let data = JSON.parse(evt.data);
        let cards = [];

        if (data.isOver) {
            //$scoreBtn.unbind();
        }

        data.scores.forEach((item) => {
            cards.push(`<li>
                <button>
                    <div class="left-icon">${item.score}</div>
                    <div class="center-icon">${item.score}</div>
                    <div class="right-icon">${item.score}</div>
                </button>
            </li>`);
        })

        $("#scores").html(cards);
    };

    ws.onclose = () => {
        console.log("Connection closed.");
        setTimeout(() => {
            ws_connect();
        }, 1000);
    };
};

window.onload = () => {
    ws_connect();
}