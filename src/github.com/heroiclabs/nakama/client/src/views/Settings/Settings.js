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


class Settings extends Component {

  constructor(props) {
    super(props);


 //   if (!dashboard.session) {
 //     dashboard.RedirectLogin();
 //   }

    this.state = {
      settings: {swagLevels: []},
      wrongCPP: false,
      wrongh2h: false,
      showSaveNot: false,
      wrongWShowup: false,
    };

    this.refresh()
  }

  refresh() {
    dashboard.getSettings((settings)=>{
      this.setState({settings: settings})
    },)
  }


  saveSettings() {
    if (this.state.wrongCPP) return;
    if (this.state.wrongh2h) return;
    if (this.state.wrongWShowup) return;

    dashboard.updateSettings(this.state.settings, (settings)=>{
      this.setState({showSaveNot: true});
      setTimeout(()=>{
        this.setState({showSaveNot: false});
      }, 2000)

    },)
  }
  isFloat(val) {
    var floatRegex = /^-?\d+(?:[.,]\d*?)?$/;
    if (!floatRegex.test(val))
      return false;

    val = parseFloat(val);
    if (isNaN(val))
      return false;
    return true;
  }



  setWordShowup(event) {
    this.state.settings.word_showup_time = event.target.value;
    if (this.isFloat(event.target.value) == false) {
      this.setState({wrongWShowup:true});
    } else {
      this.state.settings.word_showup_time = parseFloat(event.target.value);
      this.setState({wrongWShowup:false});
    }

    this.setState({settings:this.state.settings});
  }

  setH2HWords(event) {
    this.state.settings.h2hwords = event.target.value;
    if (this.isFloat(event.target.value) == false) {
      this.setState({wrongh2h:true});
    } else {
      this.state.settings.h2hwords = parseFloat(event.target.value);
      this.setState({wrongh2h:false});
    }

    this.setState({settings:this.state.settings});
  }

  setRoundLen(event) {
    this.state.settings.len_round = parseInt(event.target.value);

    this.setState({settings:this.state.settings});
  }

  setTimePenalty(event) {
    this.state.settings.time_penalty_sec = parseInt(event.target.value);

    this.setState({settings:this.state.settings});
  }

  setCoinsPerPoint(event) {
    this.state.settings.coinsPerPoint = event.target.value;
    if (this.isFloat(event.target.value) == false) {
      this.setState({wrongCPP:true});
    } else {
      this.state.settings.coinsPerPoint = parseFloat(event.target.value);
      this.setState({wrongCPP:false});
    }

    this.setState({settings:this.state.settings});
  }

  addLevel() {
    this.state.settings.swagLevels.push({
      name: "",
      points: 0,
    });

    this.setState({settings:this.state.settings});
  }

  deleteLevel(l){
    this.state.settings.swagLevels.splice(l, 1);
    this.setState({settings:this.state.settings});
  }

  setDesc(inx, event) {
    this.state.settings.swagLevels[inx].desc =  event.target.value;
    this.setState({settings:this.state.settings});
  }

  setName(inx, event) {
    this.state.settings.swagLevels[inx].name =  event.target.value;
    this.setState({settings:this.state.settings});
  }

  setPoints(inx, event) {
    this.state.settings.swagLevels[inx].points =  parseFloat(event.target.value);
    this.setState({settings:this.state.settings});
  }

  render() {
    const buttonStyle = {border: "0px"};
    const wordsList = [];
    let len = this.state.settings.swagLevels.length;
    for (let w = 0; w < len; w++) {
      wordsList.push(
        <tr key={w}>
        <td className="pt-3-half">
          <input onChange={this.setName.bind(this, w)} value={this.state.settings.swagLevels[w].name}
                 className="form-control" id="inputdefault" type="text">
          </input>
        </td>
          <td className="pt-3-half">
            <input onChange={this.setDesc.bind(this, w)} value={this.state.settings.swagLevels[w].desc}
                   className="form-control" id="inputdefault" type="text">
            </input>
          </td>
          <td  className="pt-3-half">
          <div >
            <input pattern="[0-9]*" onChange={this.setPoints.bind(this, w)} value={this.state.settings.swagLevels[w].points}
                   type="number" className="form-control" id="inputdefault" >
            </input>
          </div>
        </td>
          <td style={{"width": "5%"}}>   <button type="button" onClick={this.deleteLevel.bind(this, w)} style={buttonStyle}
                                                 className="btn btn-danger">Delete</button></td>

      </tr>)
    }

    let wrongCPP = this.state.wrongCPP?<div className="successfully-saved alert alert-danger" role="alert">
      Enter valid floating point number
    </div>:[];
    let wrongh2h = this.state.wrongh2h?<div className="successfully-saved alert alert-danger" role="alert">
      Enter valid floating point number
    </div>:[];
    let wrongWShowup = this.state.wrongWShowup?<div className="successfully-saved alert alert-danger" role="alert">
      Enter valid floating point number
    </div>:[];



    let savedNotification = this.state.showSaveNot?<div className="successfully-saved alert alert-info" role="alert">
      <strong>Category saved</strong>
    </div>:[];

    return (
      <div>
        <div className="animated fadeIn">
          <Row>
            <Col>
              <Card>
                <CardHeader>
                  Settings
                </CardHeader>
                <CardBody>
                  <form>
                    <div className="form-group">
                      <label>Coins Per Point:</label>
                      {wrongCPP}
                      <input onChange={this.setCoinsPerPoint.bind(this)} value={this.state.settings.coinsPerPoint}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Round length:</label>
                      {wrongCPP}
                      <input onChange={this.setRoundLen.bind(this)} value={this.state.settings.len_round}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Head 2 head words:</label>
                      {wrongh2h}
                      <input onChange={this.setH2HWords.bind(this)} value={this.state.settings.h2hwords}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Words showup delay (seconds):</label>
                      {wrongWShowup}
                      <input onChange={this.setWordShowup.bind(this)} value={this.state.settings.word_showup_time}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Time penalty (seconds):</label>
                      {wrongWShowup}
                      <input onChange={this.setTimePenalty.bind(this)} value={this.state.settings.time_penalty_sec}
                             className="form-control">
                      </input>
                    </div>
                  </form>
                  <br/>
                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Level Name</th>
                      <th className="text-center">Level Desc</th>
                      <th className="text-center">Points</th>
                      <th></th>
                    </tr>
                    </thead>
                    <tbody>
                    {wordsList}
                    <tr>
                      <td>
                      </td>
                      <td>
                      </td>

                      <td>
                      <input className="btn btn-primary" onClick={this.addLevel.bind(this)} style={buttonStyle}
                             type="button" value="Add level">
                      </input>
                      </td>
                   </tr>
                    </tbody>
                  </Table>
                </CardBody>
              </Card>
              <Card>
                <CardBody>
                  {savedNotification}
                  <input className="btn btn-primary" onClick={this.saveSettings.bind(this)} style={buttonStyle}
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

export default Settings;
/*        <YesNoModal ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
        </YesNoModal>*/
