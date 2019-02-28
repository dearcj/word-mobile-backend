
let client = new nakamajs.Client("defaultkey", "localhost", 7350);
// client.ssl = true;
const doJob = function (session) {
  client.rpc(session, "create_match_rpc", {label: "some label"}).then((res) => {
    console.log(res);
  });

};

let sess = window.localStorage.getItem("sessionToken");
if  (sess) {
  let session = window.nakamajs.Session.restore(sess);
  doJob(session)
} else {


  const session = client.authenticateEmail({
    email: "ziuziuziu@gmail.com",
    password: "testtest",
    create: true
  }).then((session) => {
    window.localStorage.setItem("sessionToken", session.token);
    doJob(session);
    console.info("Successfully authenticated:", session);
  });
}


