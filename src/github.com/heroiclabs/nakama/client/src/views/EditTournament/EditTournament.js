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
import Select from 'react-select';
import countryOptions from '../../countryList';
import 'react-datepicker/dist/react-datepicker.css';
import DatePicker from "react-datepicker";
import moment from 'moment';

const brandPrimary = getStyle('--primary');
const brandInfo = getStyle('--info');
const langs = ["english", "spanish", "italian", "french", "portuguese"];


class EditTournament extends Component {


  save(cb) {
    dashboard.updateTournament(this.state.tournament, this.state.deleteRounds, (res) => {
      if (res.payload.tournament.rounds)
        for (let x of res.payload.tournament.rounds) {
          this.addMoment(x)
        }

      this.setState({
        deleteRound: [],
        tournament: res.payload.tournament
      });

      if (cb) cb();

      this.setState({showSaveNot: true});
      setTimeout(() => {
        this.setState({showSaveNot: false});
      }, 2000)
    });
  }

  update() {
    dashboard.getTournament(this.state.tournament.id, (res) => {
      if (res.rounds)
        for (let x of res.rounds) {
          this.addMoment(x)
        }

      this.setState({
        start_date_moment: moment(res.start_date),
        tournament: res, location: {option: res.location, label: res.location.replace(/^\w/, c => c.toUpperCase())}
      });

      this.updateImage()
    })
  }

  constructor(props) {
    super(props);

    if (!dashboard.session) {
      dashboard.RedirectLogin();
    }

    this.state = {
      showPublNot: false,
      noRounds: false,
      tournament: dashboard.editingTournament,
      showSaveNot: false,
      countryOptions: [{label: "Any", option: "any"}],
      deleteRounds: [],
    };

    dashboard.getCountries((res)=>{
      let options = [];
      for (let x of res.payload) {
        options.push({label: x.replace(/^\w/, c => c.toUpperCase()), option: x.toLowerCase()})
      }

      this.setState({countryOptions: options})
    });

    this.update();
  }

  onCopy(text) {
    let textField = document.createElement('textarea');
    textField.innerText = text;
    document.body.appendChild(textField);
    textField.select();
    document.execCommand('copy');
    textField.remove();
  }

  changeDesc(event) {
    this.state.tournament.desc = event.target.value;
    this.setState({tournament: this.state.tournament});
  }

  changeName(event) {
    this.state.tournament.name = event.target.value;
    this.setState({tournament: this.state.tournament});
  }

  addMoment(r) {
    r.end_date_moment = moment(r.end_date);
  }

  addRound() {
    let r = {
      tournament_id: this.state.tournament.id,
      award: 0, end_date: new Date(), top_perc: 0
    };

    this.addMoment(r);

    if (!this.state.tournament.rounds)
      this.state.tournament.rounds = [];
    this.state.tournament.rounds.push(r);

    this.setState({tournament: this.state.tournament});
  }

  updateImage() {
  }

  changePartic(event) {
    this.state.tournament.participants = parseFloat(event.target.value);
    event.target.value = this.state.tournament.participants.toString();
    this.setState({tournament: this.state.tournament});
  }

  handleUploadFile(event) {
    let reader = new FileReader();

    reader.onload = (e) => {
      this.state.tournament.image_data = e.target.result;

      this.updateImage();
      this.setState({
        tournament: this.state.tournament,
      });
    };

    if (event.target.files.length > 0)
      reader.readAsDataURL(event.target.files[0]);

    this.setState({fileSize: event.target.files[0].size})
  }

  setLang(l) {
    this.state.tournament.language = l;

    this.state.tournament.language_str = langs[l];
    this.setState({
      tournament: this.state.tournament,
    });
  }

  deleteRound(inx) {
    if (this.state.tournament.rounds[inx].id) {
      this.state.deleteRounds.push(this.state.tournament.rounds[inx].id);
    }

    this.state.tournament.rounds.splice(inx, 1);
    this.setState({
      tournament: this.state.tournament,
    });
  }

  setLocation(selectedOption) {
    this.state.tournament.location = selectedOption.option;
    this.setState({
      location: selectedOption,
      tournament: this.state.tournament,
    });

    //this.setState({selectedLocation: selectedOption });
  }

  setPerc(inx, event) {
    let r = this.state.tournament.rounds[inx];
    r.top_perc = parseInt(event.target.value);
    if (r.top_perc > 100)
      r.top_perc = 100;
    if (r.top_perc < 0)
      r.top_perc = 0;
    this.setState({
      tournament: this.state.tournament,
    });
  }

  setStartDate(moment) {
    this.setState({
      start_date_moment: moment,
    });
    this.state.tournament.start_date = moment._d;
  }

  setRoundDate(inx, moment) {
    this.state.tournament.rounds[inx].end_date_moment = moment;
    this.state.tournament.rounds[inx].end_date = moment._d;
    this.setState({
      tournament: this.state.tournament,
    });
  }

  setAward(inx, event) {
    this.state.tournament.rounds[inx].award = parseInt(event.target.value);
    this.setState({
      tournament: this.state.tournament,
    });
  }

  publish() {
    if (this.state.tournament.rounds == null ||
      this.state.tournament.rounds.length == 0) {

      this.setState({noRounds: true});
      setTimeout(() => {
        this.setState({noRounds: false});
      }, 2000);

      return
    }

    this.save(() => {
      dashboard.publishTournament(this.state.tournament.id, () => {
        this.setState({showPublNot: true});
        setTimeout(() => {
          this.setState({showPublNot: false});
        }, 2000)

        this.update();
      }, () => {

      });
    });

  }

  render() {
    const buttonStyle = {border: "0px"};

    window.$('#img-upload').attr('src', this.state.tournament.image_data);

    let imageTooBig = this.state.fileSize > 120000 ?
      <div className="successfully-saved alert alert-danger" role="alert">
        <strong>Image is too big: {Math.round(this.state.fileSize / 1000.)}KB</strong>
      </div> : [];
    let savedNotification = this.state.showSaveNot ? <div className="successfully-saved alert alert-info" role="alert">
      <strong>Tournament saved</strong>
    </div> : [];
    let tournamentPublished = this.state.showPublNot ?
      <div className="successfully-saved alert alert-info" role="alert">
        <strong>Tournament published</strong>
      </div> : [];
    let noRoundsError = this.state.noRounds ? <div className="successfully-saved alert alert-danger" role="alert">
      <strong>Cannot publish tournament without rounds</strong>
    </div> : [];

    let roundsList = [];
    let inx = 0;
    let publishBtn = this.state.tournament.status == 0 ?
      <input className="btn btn-warning" onClick={this.publish.bind(this, dashboard)} style={buttonStyle} type="button"
             value="Publish"></input> : [];

    for (let r in this.state.tournament.rounds) {
      roundsList.push(<tr key={inx}>
        <td style={{"width": "10%"}} className="pt-3-half">
          <input className="btn btn-primary" style={buttonStyle} type="button" onClick={this.onCopy.bind(this, r.id)}
                 value="Copy"></input>
        </td>
        <td style={{"width": "15%"}} className="pt-3-half">

          <DatePicker
            popperModifiers={{
              offset: {
                enabled: true,
                offset: '0px, 12px'
              },
              preventOverflow: {
                enabled: true,
                escapeWithReference: false,
                boundariesElement: 'viewport'
              }
            }}
            dateFormat="M/D/YY h:mm a"
            showTimeSelect
            className={"form-control"}
            selected={this.state.tournament.rounds[inx].end_date_moment}
            onChange={this.setRoundDate.bind(this, inx)}
          />

        </td>
        <td className="pt-3-half">
          <input type="number" onChange={this.setAward.bind(this, inx)} value={this.state.tournament.rounds[inx].award}
                 className="form-control" id="inputdefault">
          </input>
        </td>
        <td style={{"width": "15%"}} className="pt-3-half">
          <input type="number" onChange={this.setPerc.bind(this, inx)}
                 value={this.state.tournament.rounds[inx].top_perc}
                 className="form-control" id="inputdefault">
          </input>
        </td>


        <td style={{"width": "5%"}} className="text-right">
          <button type="button" onClick={this.deleteRound.bind(this, inx)} style={buttonStyle}
                  className="btn btn-danger">Delete
          </button>
        </td>
      </tr>);
      inx++;
    }


    return (
      <div>
        <div className="animated fadeIn">
          <Row>
            <Col>
              <Card>
                <CardHeader>
                  Edit tournament
                </CardHeader>
                <CardBody>
                  <form>
                    <div className="form-group">
                      <label>Name:</label>
                      <input onChange={this.changeName.bind(this)} value={this.state.tournament.name}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Status:</label>
                      <input className="form-control" type="text"
                             value={this.state.tournament.status_str.replace(/^\w/, c => c.toUpperCase())} disabled>
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Participants:</label>
                      <input value={this.state.tournament.participants} pattern="[0-9]*"
                             onChange={this.changePartic.bind(this)}
                             className="form-control" type="number">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Start date:</label>
                      <style>
                        {`.react-datepicker__time-container .react-datepicker__time .react-datepicker__time-box ul.react-datepicker__time-list {
              padding-left: unset;
              padding-right: unset;
              width: 100px;
            }
              .react-datepicker__input-container, .react-datepicker-wrapper {
              width:100%;
            }
              .react-datepicker {
              width: 324px;
            }
              .react-datepicker .react-datepicker__day {
              line-height: 1.4rem;
              margin: .1rem 0.166rem;
            }
              .react-datepicker .react-datepicker__day-name {
              line-height: 1rem;
            }`}
                      </style>
                      <DatePicker
                        popperModifiers={{
                          offset: {
                            enabled: true,
                            offset: '0px, 12px'
                          },
                          preventOverflow: {
                            enabled: true,
                            escapeWithReference: false,
                            boundariesElement: 'viewport'
                          }
                        }}
                        dateFormat="M/D/YY h:mm a"
                        showTimeSelect className={"form-control"}
                        selected={this.state.start_date_moment}
                        onChange={this.setStartDate.bind(this)}
                      />
                    </div>


                    <div className="form-group">
                      <label>Description:</label>
                      <input onChange={this.changeDesc.bind(this)} value={this.state.tournament.desc}
                             className="form-control"></input>
                    </div>
                    <div className="form-group">
                      <label>Language:</label>
                      <div className="dropdown">
                        <button className="btn btn-secondary dropdown-toggle" type="button" id="dropdownMenuButton"
                                data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
                          {this.state.tournament.language_str.replace(/^\w/, c => c.toUpperCase())}
                        </button>
                        <div className="dropdown-menu">
                          <a className="dropdown-item" style={{"cursor": "pointer"}}
                             onClick={this.setLang.bind(this, 0)}>English</a>
                          <a className="dropdown-item" style={{"cursor": "pointer"}}
                             onClick={this.setLang.bind(this, 1)}>Spanish</a>
                          <a className="dropdown-item" style={{"cursor": "pointer"}}
                             onClick={this.setLang.bind(this, 2)}>Italian</a>
                          <a className="dropdown-item" style={{"cursor": "pointer"}}
                             onClick={this.setLang.bind(this, 3)}>French</a>
                          <a className="dropdown-item" style={{"cursor": "pointer"}}
                             onClick={this.setLang.bind(this, 4)}>Portuguese</a>
                        </div>
                      </div>
                    </div>
                    <div className="form-group">
                      <label>Location:</label>
                      <Select
                        value={this.state.location}
                        onChange={this.setLocation.bind(this)}
                        options={this.state.countryOptions}
                      ></Select>
                    </div>
                    <div className="form-group">
                      <label>Tournament image (jpg, 413x370px, max: 120kB)</label>
                      <div className="input-group">
            <span className="input-group-btn">
                  <input onChange={this.handleUploadFile.bind(this)} type="file" accept="image/jpeg"
                         id="imgInp"></input>
            </span>
                      </div>
                      <br/>
                      {imageTooBig}
                      <img id="img-upload" style={{"max-width": "480px"}}></img>
                    </div>
                  </form>
                  <br/>
                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Id</th>
                      <th className="text-center">End date</th>
                      <th className="text-center">Award</th>
                      <th className="text-center">Win percentage</th>
                      <th></th>
                    </tr>
                    </thead>
                    <tbody>
                    {roundsList}
                    <tr className="invalid">
                      <td>
                      </td>
                      <td className="text-center">
                      </td>
                      <td>
                      </td>
                      <td className="text-center">
                      </td>
                      <td className="text-right">
                        <input className="btn btn-primary" onClick={this.addRound.bind(this)} style={buttonStyle}
                               type="button" value="Add Round"></input>
                      </td>
                    </tr>
                    </tbody>
                  </Table>

                </CardBody>

              </Card>
              <Card>
                <CardBody className="float-right">
                  {savedNotification}
                  {noRoundsError}
                  {tournamentPublished}
                  <span className="float-right" style={{"margin-right": "15px"}}>
                  <input className="btn btn-primary " onClick={this.save.bind(this, null)} style={buttonStyle}
                         type="button" value="Save"></input>
                  </span>
                  {publishBtn}
                </CardBody>
              </Card>
            </Col>
          </Row>

        </div>

      </div>
    );
  }
}

export default EditTournament;
/*        <YesNoModal ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
        </YesNoModal>*/
