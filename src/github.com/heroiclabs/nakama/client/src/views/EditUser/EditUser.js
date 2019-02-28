import React, {Component} from 'react';
import {getStyle, hexToRgba} from '@coreui/coreui/dist/js/coreui-utilities'
import {dashboard} from '../../App';
import {
  Badge,
  Button,
  ButtonDropdown,
  ButtonGroup,
  ButtonToolbar,
  Card,
  CardBody,
  CardFooter,
  CardHeader,
  CardTitle,
  Col,
  Dropdown,
  DropdownItem,
  DropdownMenu,
  DropdownToggle,
  Progress,
  Row,
  Table,
} from 'reactstrap';

const brandPrimary = getStyle('--primary');
const brandInfo = getStyle('--info');


class EditUser extends Component {
  deleteStorage(cid, cname) {
    /*  this.setState({
        currentCatId: cid,
        currentCatName: cname,
      });*/
  }

  addStorage() {
    /*  for (let x of this.state.category.languages) {
        x.words.push("");
      }
      this.setState({category: this.state.category});*/
  }

  update() {
    dashboard.getUserWithStorage(dashboard.editingUserId, (res) => {
      this.setState({
        user: res,

      });

      this.getBoosts(res);
    });
  }

  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.onRadioBtnClick = this.onRadioBtnClick.bind(this);

    if (!dashboard.session) {
      dashboard.RedirectLogin();
    }

    this.state = {
      boosts: {list: []},
      showSaveErr: false,
      showSaveNot: false,
      user: {},
    };

    this.update(dashboard.editingUserId);
  }

  save() {
    let boosts = this.findBoosts(this.state.user);
    boosts.Value = this.encodeBoosts(this.state.boosts);
    dashboard.updateUser(this.state.user, (cat) => {

      this.setState({showSaveNot: true});
      setTimeout(() => {
        this.setState({showSaveNot: false});
      }, 2000)
    }, () => {
      this.setState({showSaveErr: true});
      setTimeout(() => {
        this.setState({showSaveErr: false});
      }, 2000)
    });
  }

  onCopy(text) {
    let textField = document.createElement('textarea');
    textField.innerText = text;
    document.body.appendChild(textField);
    textField.select();
    document.execCommand('copy');
    textField.remove();
  }

  toggle() {
    this.setState({
      dropdownOpen: !this.state.dropdownOpen,
    });
  }

  onRadioBtnClick(radioSelected) {
    this.setState({
      radioSelected: radioSelected,
    });
  }

  changeFacebookId(event) {
    this.state.user.Facebook_Id = event.target.value;
    this.setState({user: this.state.user});
  }


  changeEmail(event) {
    this.state.user.Email = event.target.value;
    this.setState({user: this.state.user});
  }

  changeUsername(event) {
    this.state.user.Username = event.target.value;
    this.setState({user: this.state.user});
  }

  getValue(v) {
    return v;
  }

  setValue(w, event) {
    this.state.user.StorageData[w].Value = event.target.value;
    this.setState({user: this.state.user});
  }

  findBoosts(u) {
    var boosts;


    for (var i = 0; i < u.StorageData.length; i++) {
      if (u.StorageData[i].Key === "boosts") {
        boosts = u.StorageData[i];
        break;
      }
    }

    return boosts;
  }

  getBoosts(u) {
    let boosts = this.findBoosts(u);
    if (boosts) {
      this.state.boosts = this.decodeBoosts(boosts.Value)
    } else {
      this.state.boosts = {
        list:
          [{type: "Time", amount: 0},
            {type: "Skip", amount: 0},
            {type: "Hint", amount: 0},
            {type: "Redo", amount: 0}]
      };

      u.StorageData.push({Collection: "resources", Key: "boosts", Read: 1, Write: 1, Value: this.encodeBoosts(this.state.boosts)});
    }

    this.setState(this.state)
  }

  changeAmount(inx, event) {
    this.state.boosts.list[inx].amount = event.target.value.toString();
    this.setState({boosts: this.state.boosts});
  }

  decodeBoosts(boostsFromStorage) {
    let decoded = decodeURIComponent(boostsFromStorage);
    let json = JSON.parse(decoded);
    return json;
  }

  encodeBoosts(boosts) {
    var stringified = JSON.stringify(boosts);
    var encoded = encodeURIComponent(stringified);
    return encoded;
  }

  render() {
    const buttonStyle = {border: "0px"};
    let boostsList = [];
    for (let i = 0; i < this.state.boosts.list.length; i++) {
      let x = this.state.boosts.list[i];
      boostsList.push(<tr>
        <td className="pt-3-half">
          <div>
            {x.type}
          </div>
        </td>
        <td className="pt-3-half">
          <input pattern="[0-9]*" onChange={this.changeAmount.bind(this, i)}
                 value={x.amount}
                 className="form-control" type="number">
          </input>
        </td>
      </tr>)
    }

    let saveError = this.state.showSaveErr ? <div className="successfully-saved alert alert-danger" role="alert">
      Could not save user, email and username must be unique
    </div> : [];

    let savedNotification = this.state.showSaveNot ? <div className="successfully-saved alert alert-info" role="alert">
      <strong>User saved</strong>
    </div> : [];
    let list = [];
    let len = this.state.user.StorageData ? this.state.user.StorageData.length : 0;
    for (let w = 0; w < len; w++) {
      let sd = this.state.user.StorageData[w];
      if (sd.Key === 'boosts') continue;
      list.push(<tr key={w}>
        <td className="pt-3-half">
          <div>
            {sd.Collection}
          </div>
        </td>
        <td className="pt-3-half">
          <div>
            {sd.Key}
          </div>
        </td>
        <td className="pt-3-half">
          <input onChange={this.setValue.bind(this, w)} value={this.getValue(sd.Value)}
                 className="form-control" id="inputdefault" type="text">
          </input>
        </td>
        <td className="pt-3-half">
          <div>
            {sd.Read == "1" ? "Yes" : "No"}
          </div>
        </td>
        <td className="pt-3-half">
          <div>
            {sd.Write == "1" ? "Yes" : "No"}
          </div>
        </td>

      </tr>)
    }


    return (
      <div>
        <div className="animated fadeIn">
          <Row>
            <Col>
              <Card>
                <CardHeader>
                  Edit user
                </CardHeader>
                <CardBody>
                  <form>
                    <div className="form-group">
                      <label>Email:</label>
                      <input onChange={this.changeEmail.bind(this)}
                             value={this.state.user.Email ? this.state.user.Email : ""}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Username:</label>
                      <input onChange={this.changeUsername.bind(this)}
                             value={this.state.user.Username ? this.state.user.Username : ""}
                             className="form-control" type="text">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Facebook Id:</label>
                      <input onChange={this.changeFacebookId.bind(this)}
                             value={this.state.user.Facebook_Id ? this.state.user.Facebook_Id : ""}
                             className="form-control" type="text">
                      </input>
                    </div>
                  </form>
                  <br/>
                  <label>User boosts:</label>
                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Type</th>
                      <th className="text-center">Amount</th>
                    </tr>
                    </thead>
                    <tbody>
                    {boostsList}
                    </tbody>
                  </Table>

                  <br></br>
                  <label>User storage:</label>
                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Collection</th>
                      <th className="text-center">Key</th>
                      <th className="text-center">Value</th>
                      <th className="text-center">Can Read</th>
                      <th className="text-center">Can Write</th>
                    </tr>
                    </thead>
                    <tbody>
                    {list}

                    </tbody>
                  </Table>

                </CardBody>

              </Card>
              <Card>
                <CardBody>
                  {savedNotification}
                  {saveError}
                  <input className="btn btn-primary" onClick={this.save.bind(this)} style={buttonStyle}
                         type="button" value="Save"></input>
                </CardBody>
              </Card>
            </Col>
          </Row>

        </div>

      </div>
    );
  }

}

export default EditUser;
/*        <YesNoModal ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
        </YesNoModal>*/
