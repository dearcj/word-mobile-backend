const HTTP_KEY = "euM9IKWVv4ESY3THvtSWzACM43vjwfaW";
let client = new nakamajs.Client(HTTP_KEY, "localhost", 7350);

let client2 = () =>{
  window.document.getElementById("send").addEventListener("click", ()=>{
    client.rpc(window.session, "send_invite", {payload: {opponent_id: "3a76b16a-fae6-4378-a07c-1c43d7b9bf1e"}}).then((res) => {
      console.log(res);
    });
  });


  window.document.getElementById("acceptInvite").addEventListener("click", ()=>{
    client.rpc(window.session, "resolve_invite", {payload: {resolution: "accept", match_id: window.matches[0].Id, invite_id: window.matches[0].InviteId}}).then((res) => {
        console.log(res);
      });

  });

  window.document.getElementById("declineInvite").addEventListener("click", ()=>{
    client.rpc(window.session, "resolve_invite", {payload: {resolution: "decline", match_id: window.matches[0].Id, invite_id: window.matches[0].InviteId}}).then((res) => {
      console.log(res);
    });
  });

  window.document.getElementById("submit").addEventListener("click", ()=>{
    client.rpc(window.session, "resolve_h2h_round", {payload: {match_id: window.matches[0].Id, added_points: 200 + Math.random()*150}}).then((res) => {
      console.log(res);
    });
  });

  window.document.getElementById("giveup").addEventListener("click", ()=>{
    client.rpc(window.session, "give_up", {payload: {match_id: window.matches[0].Id, added_points: 200 + Math.random()*150}}).then((res) => {
      console.log(res);
    });
  });

  window.document.getElementById("lbsubmit").addEventListener("click", ()=>{
    client.writeLeaderboardRecord(window.session, "lb_month", {score: Math.round(Math.random() * 100)}).then((res) => {
      console.log(res);
    });

    client.writeLeaderboardRecord(window.session, "lb_today", {score: Math.round(Math.random() * 100)}).then((res) => {
      console.log(res);
    });

    client.writeLeaderboardRecord(window.session, "lb_week", {score: Math.round(Math.random() * 100)}).then((res) => {
      console.log(res);
    });
  });



  const session = client.authenticateEmail({email: "inviter@gmail.com", password: "admin123", create: true }).then((session) => {
    let token = session.token;
    let userid = session.user_id;
    console.log(session);

    client.listLeaderboardRecords(session, "lb_today").then((result)=>{
      for (let x of result.records) {
        console.log("Record username %o and score %o", x.username, x.score);
      }
    });

    let socket = client.createSocket(false, true);
    socket.connect(session).then(()=>{});
    socket.onnotification = (not)=>{
      console.log("SOME NOTIFICATION");
      console.log(not);
    };
    window.session = session;
    client.rpc(session, "get_opponents", {payload: ""}).then((res) => {
      window.matches =   res.payload.Matches;
      console.log(res.payload);
    });

    client.rpc(session, "settings", {payload: ""}).then((res) => {
      console.log(res.payload);
    });

    client.rpc(session, "track_purchase", {payload: {items: [{
          receipt: "zzzzzzzz",
          amount: 1,
          item: "booster_1"
        },
        {
          receipt: "zzzzzzzz",
          amount: 1,
          item: "booster_1"
        }]}}).then((res) => {
      console.log(res.payload);
    });


    console.info("Successfully authenticated:", session);
  });
};


client2();
