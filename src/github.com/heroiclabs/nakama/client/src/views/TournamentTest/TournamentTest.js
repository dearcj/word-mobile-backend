import React, {Component} from 'react';
import {getStyle, hexToRgba} from '@coreui/coreui/dist/js/coreui-utilities'
import {dashboard} from '../../App';
import moment from 'moment';
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


class TournamentTest extends Component {

  constructor(props) {
    super(props);

    if (!dashboard.session) {
      dashboard.RedirectLogin();
    }

    this.state = {
      availTournaments: [],
      myTournaments: [],
      submitValue: 100,};

  }

  changeSubmitValue(e) {
    this.setState({submitValue: parseFloat(e.target.value)});
  }

  getAvail() {
    dashboard.nakamaclient.rpc(dashboard.session, "get_available_tournaments", {payload: {}}).then((res) => {
      console.log(res);
      this.setState({availTournaments: res.payload.tournaments ? res.payload.tournaments : []})
    })
  }

  getActive() {
    dashboard.nakamaclient.rpc(dashboard.session, "get_current_tournaments", {payload: {}}).then((res) => {
      console.log(res);
      this.setState({myTournaments: res.payload.tournaments? res.payload.tournaments : []})
    })
  }

  submitScore(inx) {
    let tid = this.state.myTournaments[inx].id;
    dashboard.nakamaclient.rpc(dashboard.session, "tournament_make_turn", {payload: {tournament_id: tid, points: this.state.submitValue}}).then((res) => {
      console.log(res);
      this.setState({myTournaments: res.payload.tournaments? res.payload.tournaments : []})
    })
  }

  leave(inx) {
    let tid = this.state.myTournaments[inx].id;
    dashboard.nakamaclient.rpc(dashboard.session, "leave_tournament", {payload: {tournament_id: tid}}).then((res) => {
      console.log(res);
      this.setState({myTournaments: res.payload.tournaments? res.payload.tournaments : []})
    })
  }

  fin(inx) {
    let tid = this.state.myTournaments[inx].id;
    dashboard.nakamaclient.rpc(dashboard.session, "finish_tournament", {payload: {tournament_id: tid}}).then((res) => {
      console.log(res);
      this.setState({myTournaments: res.payload.tournaments? res.payload.tournaments : []})
    })
  }

  forward(id) {
    dashboard.nakamaclient.rpc(dashboard.session, "dashboard_tournament_forward", {payload: {tournament_id: id}}).then((res) => {
      console.log(res);
    })
  }

  join(inx) {
    let tid = this.state.availTournaments[inx].id;
    dashboard.nakamaclient.rpc(dashboard.session, "join_tournament", {payload: {tournament_id: tid}}).then((res) => {
      console.log(res);
      this.setState({myTournaments: res.payload.tournaments? res.payload.tournaments : []})
    })
  }

  render() {
    const buttonStyle = {border: "0px"};

    let tournaments = [];
    let inx = 0;
    for (let t of this.state.availTournaments) {
      tournaments.push(<tr key={t.id}>
        <td style={{"width": "5%"}} >
          {t.id}
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          {t.name}
        </td>
        <td className="pt-3-half">
          {t.status_str}
        </td>
        <td className="pt-3-half">
          {t.participants}
        </td>
        <td className="pt-3-half">
          {t.cur_participants}
        </td>
        <td className="pt-3-half">
          {moment(t.start_date).format('MMMM Do YYYY, h:mm:ss a')}
        </td>
        <td className="pt-3-half">
          <input className="btn btn-primary" onClick={this.join.bind(this, inx)} style={buttonStyle}
                 type="button" value="Join"></input>        </td>
        <td>  <input className="btn btn-primary" onClick={this.forward.bind(this, t.id)} style={buttonStyle}
                     type="button" value="Forward"></input>        </td>
      </tr>);
      inx++
    }

    inx = 0;
    let myt = [];
    for (let t of this.state.myTournaments) {
      myt.push(<tr key={t.id}>
        <td style={{"width": "5%"}} >
          {t.id}
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          {t.name}
        </td>
        <td className="pt-3-half">
          {t.status_str}
        </td>
        <td className="pt-3-half">
          {t.participants}
        </td>
        <td className="pt-3-half">
          {t.cur_participants}
        </td>
        <td className="pt-3-half">
        <input className="btn btn-primary" onClick={this.submitScore.bind(this, inx)} style={buttonStyle}
               type="button" value="Submit"></input>        </td>
        <td className="pt-3-half">
          <input className="btn btn-primary" onClick={this.leave.bind(this, inx)} style={buttonStyle}
                 type="button" value="Leave"></input>        </td>
        <td>  <input className="btn btn-primary" onClick={this.fin.bind(this, inx)} style={buttonStyle}
                 type="button" value="Finish"></input>        </td>
        <td>  <input className="btn btn-primary" onClick={this.forward.bind(this, t.id)} style={buttonStyle}
                     type="button" value="Forward"></input>        </td>
        </tr>);
      inx++
    }

    return (
      <div>
        <div className="animated fadeIn">
          <Row>
            <Col>
              <Card>
                <CardHeader>
                  Test

                </CardHeader>
                <CardBody>
                  <div className="form-check">
                    <input value={this.state.submitValue} onChange={this.changeSubmitValue.bind(this)}  type="text" className="form-check-input" id="unpub-input" ></input>
                    <label className="form-check-label" htmlFor="unpub-input">Submit value</label>
                  </div>
                  <input className="btn btn-primary" onClick={this.getAvail.bind(this)} style={buttonStyle}
                         type="button" value="GetAvailTournaments"></input>
                  <br/>
                  <input className="btn btn-primary" onClick={this.getActive.bind(this)} style={buttonStyle}
                         type="button" value="GetActiveTournaments"></input>
                  <br/>
                  <br/>
                  Avail tournaments
                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Id</th>
                      <th className="text-center">Name</th>
                      <th className="text-center">Status</th>
                      <th className="text-center">Max Part.</th>
                      <th className="text-center">Part</th>
                      <th className="text-center">Start date</th>
                      <th></th>
                      <th></th>
                      <th></th>
                      <th></th>
                    </tr>
                    </thead>
                    <tbody>
                    {tournaments}

                    </tbody>
                  </Table>
                  <br/>

                  My tournament
                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Id</th>
                      <th className="text-center">Name</th>
                      <th className="text-center">Status</th>
                      <th className="text-center">Max Part.</th>
                      <th className="text-center">Part</th>
                      <th></th>
                      <th></th>
                      <th></th>
                      <th></th>
                    </tr>
                    </thead>
                    <tbody>
                    {myt}

                    </tbody>
                  </Table>

                </CardBody>

              </Card>
            </Col>
          </Row>

        </div>

      </div>
    );
  }
}

export default TournamentTest;
/*        <YesNoModal ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
        </YesNoModal>*/
