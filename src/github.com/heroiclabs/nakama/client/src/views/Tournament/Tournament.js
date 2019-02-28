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
import {YesNoModal} from "../Dashboard";

const brandPrimary = getStyle('--primary');
const brandInfo = getStyle('--info');


class Tournament extends Component {
  toDelete(t) {
    this.setState({curTournament: t})
  }

  deleteTournament(cid) {
     dashboard.deleteTournament(this.state.curTournament.id, (res)=>{
       this.update();
       console.log(res);
     });
  }

  updateTournament() {
    dashboard.updateTournament(this.state.tournament, (cat) => {
      //dashboard.editingCategory = cat.payload;
      this.update(this.state.category.id);


      this.setState({showSaveNot: true});
      setTimeout(()=>{
        this.setState({showSaveNot: false});
      }, 2000)
    });
  }

  update() {
    dashboard.getTournaments(this.state.filter, (tournaments) => {
      //dashboard.editingCategory = cat.payload;
      this.setState({tournaments: tournaments?tournaments:[]})

    });
  }


  save() {

  }

  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.onRadioBtnClick = this.onRadioBtnClick.bind(this);



    this.state = {
      curTournament: {},
      tournaments: [],
      currentCat: null,
      currentCatId: "",
      currentCatName: "",
      showSaveNot: false,
      filter: [false,false,false,false],
    };
    if (!dashboard.session) {
      dashboard.RedirectLogin();
    } else {
      this.update()
    }
  //  this.update(this.state.category.id);
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


  editTournament(t) {
    dashboard.editingTournament = t;
    window.location = '#edit_tournament';
  }

  addTournament() {
    dashboard.addTournament(() => {
      this.update();
    });
  }

  changeUnpublished(e){
    this.state.filter[0] = e.target.checked;
    this.update()
  }

  changePublished(e){
    this.state.filter[1] = e.target.checked;
    this.update()
  }

  changeActive(e){
    this.state.filter[2] = e.target.checked;
    this.update()
  }

  changeFinished(e){
    this.state.filter[3] = e.target.checked;
    this.update()
  }
  onCopy(text) {
    let textField = document.createElement('textarea');
    textField.innerText = text;
    document.body.appendChild(textField);
    textField.select();
    document.execCommand('copy');
    textField.remove();
  }

  summRounds(inx) {
    let summ = 0;
    for (let x in this.state.tournaments[inx].rounds) {
      summ += this.state.tournaments[inx].rounds[x].award
    }

    return summ
  }

  handleUploadFile(event) {
    let reader = new FileReader();

    reader.onload = (e)=> {
      this.state.category.imageData = e.target.result;

      this.setState({
        category: this.state.category,
      });
    };

    if (event.target.files.length > 0)
    reader.readAsDataURL(event.target.files[0]);

    this.setState({fileSize: event.target.files[0].size})
  }

  render() {
    const buttonStyle = {border: "0px"};

    let tournaments = [];
    let inx = 0;
    for (let t of this.state.tournaments) {
      tournaments.push(<tr key={t.id}>
        <td style={{"width": "5%"}} >
          <input className="btn btn-primary" style={buttonStyle} type="button" onClick={this.onCopy.bind(this, t.id)} value="Copy"></input>
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          {t.name}
        </td>
        <td className="pt-3-half">
          {t.location.replace(/^\w/, c => c.toUpperCase())}
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
            {t.language_str.replace(/^\w/, c => c.toUpperCase())}
        </td>
        <td className="pt-3-half">
          {moment(t.start_date).format('LL')}
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          {t.cur_participants} / {t.participants}
        </td>
        <td className="pt-3-half">
          {t.rounds?t.rounds.length:0}
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          {this.summRounds(inx)}
        </td>
        <td className="pt-3-half">
          {t.status_str.replace(/^\w/, c => c.toUpperCase())}
        </td>
        <td style={{"width": "3%"}} className="text-right">
        <span style={buttonStyle} onClick={this.editTournament.bind(this, t)} type="button" className="btn btn-default btn-sm">
            <span className="cui-pencil h5"></span>
          </span>
        </td>
        <td style={{"width": "5%"}} className="text-right">
          <button type="button" onClick={this.toDelete.bind(this, t)} style={buttonStyle}
                  data-toggle="modal" data-target="#yesnomodal" style={buttonStyle}
                  className="btn btn-danger">Delete
          </button>
        </td>
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
                  Tournaments
                </CardHeader>
                <CardBody>
                  Filter:
                  <div className="form-check">
                    <input onChange={this.changeUnpublished.bind(this)} type="checkbox" className="form-check-input" id="unpub-input" ></input>
                      <label className="form-check-label" htmlFor="unpub-input">Unpublished</label>
                  </div>
                  <div className="form-check">
                    <input onChange={this.changePublished.bind(this)}  id="pub-input" type="checkbox" className="form-check-input" ></input>
                    <label className="form-check-label" htmlFor="pub-input">Published</label>
                  </div>
                  <div className="form-check">
                    <input onChange={this.changeActive.bind(this)} id="active-input" type="checkbox" className="form-check-input" ></input>
                    <label className="form-check-label" htmlFor="active-input">Active</label>
                  </div>
                  <div className="form-check">
                    <input onChange={this.changeFinished.bind(this)} id="fin-input" type="checkbox" className="form-check-input" ></input>
                    <label className="form-check-label" htmlFor="fin-input">Finished</label>
                  </div>

                  <br/>
                  <br/>
                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Id</th>
                      <th className="text-center">Name</th>
                      <th className="text-center">Location</th>
                      <th className="text-center">Language</th>
                      <th className="text-center">Start date</th>
                      <th className="text-center">Max participants</th>
                      <th className="text-center">Rounds</th>
                      <th className="text-center">Award</th>
                      <th className="text-center">Status</th>
                      <th></th>
                      <th></th>
                    </tr>
                    </thead>
                    <tbody>
                    {tournaments}
                    <tr className="invalid">
                      <td>
                      </td>
                      <td className="text-center">
                      </td>
                      <td>
                      </td>
                      <td className="text-center">
                      </td>
                      <td>
                      </td>
                      <td>
                      </td>
                      <td>
                      </td>
                      <td>
                      </td>
                      <td>
                      </td>
                      <td>
                      </td>
                      <td className="text-right">
                        <input className="btn btn-primary" onClick={this.addTournament.bind(this)} style={buttonStyle}
                               type="button" value="Add"></input>
                      </td>
                    </tr>
                    </tbody>
                  </Table>

                </CardBody>

              </Card>
            </Col>
          </Row>

        </div>
        <YesNoModal ondelete={this.deleteTournament.bind(this)}  nottext={"Are you sure you want to delele tournament '" + this.state.curTournament.id + "'?"} catid={this.state.currentCatId} catname={this.state.curTournament.name}>
        </YesNoModal>
      </div>
    );
  }
}

export default Tournament;
/*        <YesNoModal ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
        </YesNoModal>*/
