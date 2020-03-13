let wsUri;
let output;
let myuid;
let myid;
let clientid;
let uid;
let chan;
let cmd;
let data;
let ws;
let seq = 1001;
let profile;
let count;
let _count = 0;

function genseq() {
  return seq++;
}

window.addEventListener("load", function(evt) {
  wsUri  = "ws://" + window.location.host + "/stream";
  output = document.getElementById("output");
  myuid = document.getElementById("myuid");
  myuid.value = parseInt(1000 + Math.random() * 100);
  count = document.getElementById("count");
  clientid = document.getElementById("clientid");
  uid = document.getElementById("uid");
  chan = document.getElementById("chan");
  cmd  = document.getElementById("cmd");
  data  = document.getElementById("data");

  let print = function (message) {
    let d = document.createElement("div");
    d.innerHTML = message;
    output.appendChild(d);
  };

  let parse = function (evt) {
    JSON.parse(evt.data);
  };

  let newSocket = function () {
    let ws = new WebSocket(wsUri);
    let login = function () {
      let req = {
        cmd: "login",
        seq: genseq(),
        immed: true,
        data: {
          uid: parseInt(myuid.value),
        }
      };
      let s = JSON.stringify([req]);
      // console.log("send:", s);
      ws.send(s);
      profile = +new Date();
    };

    ws.onopen = function (evt) {
      //print('<span style="color: green;">OnOpen</span>');
      login();
    };
    ws.onclose = function (evt) {
      //print('<span style="color: red;">OnClose</span>');
      ws = null;
    };
    ws.onmessage = function (evt) {
      //print('<span style="color: blue;">OnMessage</span>');
      //console.log("recv("+(+new Date()-profile)+"ms):", evt.data);
      count.value = ++_count;
      let data = JSON.parse(evt.data);
      for (let i in data) {
        let rsp = data[i];
        switch (rsp.cmd) {
          case "login":
            //myid.value = rsp.data.id;
            //print('<span style="color: blue;">Logged in</span>');
            break;
          case "rcvdata":
            //print('<span style="color: blue;">['+rsp.data.id+':'+rsp.data.uid+':'+rsp.data.chan+'] '+rsp.data.data+'</span>');
            break;
          default:
            //print('<span style="color: blue;"></span>');;
        }
      }
    };
    ws.onerror = function (evt) {
      //print('<span style="color: red;">OnError</span>');
    };
    return ws;
  };

  for (i=0;i<1000;i++) {
    let ws = newSocket();
  }
});
