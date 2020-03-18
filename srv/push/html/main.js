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

function genseq() {
  return seq++;
}

window.addEventListener("load", function(evt) {
  wsUri  = "ws://" + window.location.host + "/stream";
  output = document.getElementById("output");
  myuid = document.getElementById("myuid");
  myuid.value = parseInt(1000 + Math.random() * 100);
  myid = document.getElementById("myid");
  clientid = document.getElementById("clientid");
  uid = document.getElementById("uid");
  uid.value = myuid.value
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
    ws = new WebSocket(wsUri);
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
      console.log("send:", s);
      ws.send(s);
      profile = +new Date();
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
      //print('<span style="color: blue;">OnMessage</span>');
      console.log("recv("+(+new Date()-profile)+"ms):", evt.data);
      let data = JSON.parse(evt.data);
      for (let i in data) {
        let rsp = data[i];
        switch (rsp.cmd) {
          case "login":
            myid.value = rsp.data.id;
            print('<span style="color: blue;">Logged in</span>');
            break;
          case "rcvdata":
            print('<span style="color: blue;">['+rsp.data.id+':'+rsp.data.uid+':'+rsp.data.chan+'] '+rsp.data.data+'</span>');
            break;
          default:
            //print('<span style="color: blue;"></span>');;
        }
      }
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
      cmd: cmd.value,
      seq: genseq(),
    };
    switch (req.cmd) {
      case "enter":
      case "exit":
        req.data = {chans: data.value.split(',')};
        break;
      case "snd2cli":
        req.data = {ids: clientid.value.split(',').map(Number), data: data.value};
        break;
      case "snd2usr":
        req.data = {uids: uid.value.split(',').map(Number), data: data.value};
        break;
      case "snd2chan":
        req.data = {chans: chan.value.split(','), data: data.value};
        break;
      default:
    }
    let s = JSON.stringify([req]);
    //print('<span style="color: blue;">Sent request: </span>' + s);
    console.log("send:", s);
    ws.send(s);
    profile = +new Date();

    return false;
  };

  document.getElementById("cancel").onclick = function(evt) {
    if (!ws) {
      return false;
    }
    ws.close();
    //print('<span style="color: red;">Request Canceled</span>');
    return false;
  };

  document.getElementById("open").onclick = function(evt) {
    if (!ws) {
      newSocket()
    }
    return false;
  };
});
