let wsUri;
let output;
let count;
let ws;
let seq = 1001;

function genseq() {
  return seq++;
}

window.addEventListener("load", function(evt) {
  wsUri  = "ws://" + window.location.host + "/push";
  output = document.getElementById("output");
  count  = document.getElementById("count");

  let print = function (message) {
    let d = document.createElement("div");
    d.innerHTML = message;
    output.appendChild(d);
  };

  let parse = function (evt) {
    JSON.parse(evt.data);
  };

  let newSocket = function () {
    ws = new WebSocket(wsUri);
    let login = function () {
      let req = {
        cmd: "login",
        seq: genseq(),
        immed: true,
        data: {
          uid: 1001
        }
      };
      let s = JSON.stringify([req]);
      console.log("send:", s);
      ws.send(s);
    };

    ws.onopen = function (evt) {
      print('<span style="color: green;">OnOpen</span>');
      login();
    };
    ws.onclose = function (evt) {
      print('<span style="color: red;">OnClose</span>');
      ws = null;
    };
    ws.onmessage = function (evt) {
      print('<span style="color: blue;">OnMmesage</span>');
      console.log("recv:", evt.data);
    };
    ws.onerror = function (evt) {
      print('<span style="color: red;">OnError</span>');
    };
  };

  newSocket();

  document.getElementById("send").onclick = function(evt) {
    if (!ws) {
      return false;
    }

    let req = {
      cmd: "hello",
      seq: genseq(),
      data: {
        name: "t5w0rd"
      }
    };
    let s = JSON.stringify([req]);
    print('<span style="color: blue;">Sent request: </span>' + s);
    console.log("send:", s);
    ws.send(s);

    return false;
  };

  document.getElementById("cancel").onclick = function(evt) {
    if (!ws) {
      return false;
    }
    ws.close();
    print('<span style="color: red;">Request Canceled</span>');
    return false;
  };

  document.getElementById("open").onclick = function(evt) {
    if (!ws) {
      newSocket()
    }
    return false;
  };
});
