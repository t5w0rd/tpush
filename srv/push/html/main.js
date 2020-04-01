let wsUri;
let output;
let myuid;
let myid;
let clientid;
let uid;
let chan;
let cmd;
let data;
let profile;


window.addEventListener("load", function(evt) {
  wsUri  = "ws://" + window.location.host + "/stream";
  output = document.getElementById("output");
  myuid = document.getElementById("myuid");
  myuid.value = parseInt(1000 + Math.random() * 100);
  myid = document.getElementById("myid");
  clientid = document.getElementById("clientid");
  uid = document.getElementById("uid");
  uid.value = myuid.value;
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

  let seq = 1001;
  let genseq = function () {
    return seq++;
  };

  let pingTimeout = 5000;
  let recvTimeout = pingTimeout + 2000;
  let pingTimer = setTimeout("ping()", pingTimeout);
  let recvTimer = setTimeout("close()", recvTimeout);

  let newClient = function () {
    let client = {};
    let conn = new WebSocket(wsUri);
    client.conn = conn;

    let send = function (s) {
      console.log("send:", s);
      conn.send(s);
      profile = +new Date();
      if (pingTimer) {
        clearTimeout(pingTimer);
      }
      pingTimer = setTimeout(ping, pingTimeout);
    };
    client.send = send;

    let close = function () {
      conn.close();
    };
    client.close = close;

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
      send(s);
    };

    let ping = function () {
      let req = {
        cmd: "ping",
        seq: genseq()
      };
      let s = JSON.stringify([req]);
      send(s);
    };

    conn.onopen = function (evt) {
      print('<span style="color: green;">OnOpen</span>');
      login();
    };
    conn.onclose = function (evt) {
      print('<span style="color: red;">OnClose</span>');
      conn = null;
      clearTimeout(pingTimer);
      clearTimeout(recvTimer);
    };
    conn.onmessage = function (evt) {
      //print('<span style="color: blue;">OnMessage</span>');
      clearTimeout(recvTimer);
      recvTimer = setTimeout(close, recvTimeout);
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
          case "ping":
            //print('<span style="color: blue;">Ping</span>');
            break;
          default:
            //print('<span style="color: blue;"></span>');;
        }
      }
    };
    conn.onerror = function (evt) {
      print('<span style="color: red;">OnError</span>');
    };
    return client;
  };

  let client = newClient();

  document.getElementById("send").onclick = function(evt) {
    if (!client) {
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
    client.send(s);

    return false;
  };

  document.getElementById("cancel").onclick = function(evt) {
    if (!client) {
      return false;
    }
    client.close();
    //print('<span style="color: red;">Request Canceled</span>');
    return false;
  };

  document.getElementById("open").onclick = function(evt) {
    if (!client) {
      client = newClient();
    }
    return false;
  };
});
