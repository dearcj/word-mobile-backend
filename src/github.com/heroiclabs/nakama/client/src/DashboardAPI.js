import {dashboard} from "./App";


const HOST = window.location.hostname;
const PORT = 7350;
const HTTP_KEY = "euM9IKWVv4ESY3THvtSWzACM43vjwfaW";

export class DashboardAPI {
  nakamaclient;
  session;

  getAchievements(cb) {
    this.nakamaclient.rpc(this.session, "get_achievements", {}).then((res) => {
      cb(res.payload);
    });
  }

  RedirectLogin() {
    window.location = '#login';
  }

  getSettings(cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_get_settings", {payload: {}}).then((res) => {
      console.log(res);
      cb(res.payload);
    })
  }

  updateSettings(payload, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_update_settings", {payload: payload}).then((res) => {
      console.log(res);
      cb(res);
    })
  }

  getUsers(page, pageSize, filterEmail, filterId, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_get_users", {payload: {pageSize: pageSize, page: page,Filter_email: filterEmail, Filter_id: filterId}}).then((res) => {
      console.log(res);
      cb(res);
    })
  }

  getUserWithStorage(id, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_get_user_with_storage", {payload: {user_id: id}}).then((res) => {
      cb(res.payload);
    })
  }


  updateCategory(cat, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_update_category", {payload: {category: cat}}).then((res) => {
      this.addClientDataCategory(res.payload);
      cb(res);
    })
  }

  deleteCategory(cid, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_delete_category", {payload: {category_id: cid}}).then((res) => {
      cb(res);
    })
  }

  deleteUser(uid, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_delete_user", {payload: {user_id: uid}}).then((res) => {
      cb(res);
    })
  }

  addCategory(cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_add_category", "").then((res) => {
      this.addClientDataCategory(res.payload);
      cb(res);
    })
  }

  getTournaments(filter, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_get_tournaments", {payload: {filter: filter}}).then((res) => {
      cb(res.payload.Tournaments);
    })
  }

  getTournament(id, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_get_tournament", {payload: {tournament_id: id}}).then((res) => {
      cb(res.payload.Tournaments[0]);
    })
  }


  deleteTournament(tid, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_delete_tournament", {payload: {tournament_id: tid}}).then((res) => {
      cb(res);
    })
  }

  addTournament(cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_add_tournament", {payload: {}}).then((res) => {
      cb(res);
    })
  }


  updateTournament(t, dr, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_update_tournament", {payload: {tournament: t, delete_rounds: dr}}).then((res) => {
      cb(res);
    })
  }

  addClientDataCategory(x) {
    x.langMap = {};
    for (let y of x.languages) {
      if (!y.words) y.words = [];
      if (!y.points) y.points = [];
      x.langMap[y.name] = y;
    }
  }

  getCategories(cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_categories", "").then((res) => {
      let cats = res.payload.categories ? res.payload.categories : [];
      for (let x of cats) {
        this.addClientDataCategory(x);
      }

      cb(cats);
    })
  }

  getCategory(id, cb) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_get_category", {payload: {category_id: id}}).then((res) => {
      this.addClientDataCategory(res.payload);
      cb(res.payload);
    })
  }

  updateUser(data, cb, err) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_update_user", {payload: {User: data}}).then((res) => {
      cb(res);
    }).catch((res)=>{
      err(res)
    });
  }

  RedirectDashboard() {
    window.location = '#categories';
    console.log("Redirected 2 dashboard");


    /*this.nakamaclient.readStorageObjects(this.session, {
      "object_ids": [{
        "collection": "resources",
        "key": "money",
        "user_id": this.session.user_id
      }]
    }).then((c)=>{
      console.log(c);
    });*/

   //  this.nakamaclient.rpc(this.session, "dashboard_categories", "").then((res) => {
   //     console.log(res);
   //   })
//      this.nakamaclient.rpc(this.session, "resolve_round", {payload: {addedPoints: 120}}).then((res) => {
//          console.log(res);
//      });


  //  this.nakamaclient.rpc(this.session, "send_invite", {payload: {opponent_id: "824e30d1-a3ee-477f-952b-b15501eaf891"}}).then((res) => {


  //    console.log(res);
  //  });

   // this.nakamaclient.rpc(this.session, "get_opponents", {payload: ""}).then((res) => {
  //    console.log(res);
  //  });
    //this.nakamaclient.rpc(this.session, "unlock_category", {payload: {category_id: 'edcc5d56-6dca-4394-94df-ce7a2c229f0c'}}).then((res) => {
    // console.log(res);
    //})
  }

  getCountries(cb) {
    dashboard.nakamaclient.rpc(this.session, "country_list", {}).then((res) => {
      cb(res);
    })
  }

  publishTournament(id, cb, err) {
    dashboard.nakamaclient.rpc(this.session, "dashboard_publish_tournament", {payload: {tournament_id: id}}).then((res) => {
      cb(res);
    }).catch((res)=>{
      err(res)
    });
  }
  Login(email, password, errCB) {
    console.log("trying to login to nakama server");


    this.nakamaclient.authenticateEmail({create: false,  email: email, password: password}).then((session)=>{
      this.session = session;



      window.localStorage.setItem("sessionToken", session.token);
      this.RedirectDashboard();
      console.log(session);
    }).catch((err, a)=>{
      if (err.text) {
      err.text().then( errorMessage => {
        let errobj = JSON.parse(errorMessage);
        errCB(errobj.error)
      })
      }
    })
  }

  constructor() {
    let ssl = window.location.protocol === "https:";
    this.nakamaclient = new window.nakamajs.Client(HTTP_KEY, HOST, PORT, ssl);

    console.log('connected to nakama [', HOST, ']:', PORT.toString());
    let sess = window.localStorage.getItem("sessionToken");
    if  (sess) {
      this.session = window.nakamajs.Session.restore(sess);
      const nowUnixEpoch = Math.floor(Date.now() / 1000);
      if (this.session.isexpired(nowUnixEpoch)) {
        window.localStorage.removeItem("sessionToken");
        console.log("Session has expired. Must reauthenticate!");
        this.RedirectLogin();
      } else {
        this.RedirectDashboard();
      }
    }

  }

}
