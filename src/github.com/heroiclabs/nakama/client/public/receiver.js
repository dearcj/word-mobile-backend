const HTTP_KEY = "testtest";
let client = new nakamajs.Client(HTTP_KEY, "localhost", 7350);


let client2 = () =>{

  window.document.getElementById("acceptInvite").addEventListener("click", ()=>{
    client.rpc(window.session, "resolve_invite", {payload: {resolution: "accept", match_id: window.matches[0].Id}}).then((res) => {
        console.log(res);
      });

  });

  window.document.getElementById("declineInvite").addEventListener("click", ()=>{
    client.rpc(window.session, "resolve_invite", {payload: {resolution: "decline", match_id: window.matches[0].Id}}).then((res) => {
      console.log(res);
    });
  });

  window.document.getElementById("submit").addEventListener("click", ()=>{
    client.rpc(window.session, "resolve_h2h_round", {payload: {match_id: window.matches[0].Id, added_points: 200}}).then((res) => {
      console.log(res);
    });
  });

  window.document.getElementById("giveup").addEventListener("click", ()=>{
    client.rpc(window.session, "give_up", {payload: {match_id: window.matches[0].Id, added_points: 200 + Math.random()*150}}).then((res) => {
      console.log(res);
    });
  });
  const session = client.authenticateEmail({email: "admin@gmail.com", password: "admin123", create: true }).then((session) => {
    let token = session.token;
    let userid = session.user_id;
    console.log(session);

    let socket = client.createSocket(false, true);
    socket.connect(session).then(()=>{});
    socket.onnotification = (not)=>{
      console.log("SOME NOTIFICATION");
      console.log(not);
    };
    window.session = session;
    client.rpc(session, "get_opponents", {payload: ""}).then((res) => {
      window.matches =   res.payload.Matches;
      if (res.payload.Matches) {
        let inviteId = res.payload.Matches[0].InviteId;
        console.log(res.payload);



      }
    });

    client.rpc(session, "settings", {payload: ""}).then((res) => {
        console.log(res.payload);


    });

    console.info("Successfully authenticated:", session);
  });
};


client2();
