const HOST = window.location.hostname;

let client = new nakamajs.Client("defaultkey", HOST, 7350);


setTimeout(()=>{
  window.document.getElementById("login").addEventListener("click", ()=>{
    window.FB.login(function(response){
      console.log(response);
      window.accessToken = response.authResponse.accessToken;
      window.signedRequest = response.authResponse.signedRequest;
      client.authenticateFacebook({token: window.accessToken, create: true, username: response.authResponse.userID, import: true }).then((cc)=>{
        console.log(cc)
      }).catch((cc)=>{
        console.log(cc)
      })

    }, {
      scope: 'user_friends,email',
      return_scopes: true
    });
  });

}, 1000);
