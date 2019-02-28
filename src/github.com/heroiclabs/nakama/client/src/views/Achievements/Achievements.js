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


class Achievements extends Component {
  update() {
    dashboard.getAchievements((res) => {
      console.log(res)
      this.setState({
        start_date_moment: moment(res.start_date),
        achievements: res.Achievements});
    })
  }

  constructor(props) {
    super(props);

    if (!dashboard.session) {
      dashboard.RedirectLogin();
    }

    this.state = {
      achievements : [],
    };

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

  render() {
    let roundsList = [];
    let inx = 0;
    const buttonStyle = {border: "0px"};

    for (let r in this.state.achievements) {
      roundsList.push(<tr key={inx}>
        <td style={{"width": "20%"}} className="pt-3-half">
          <input className="btn btn-primary" style={buttonStyle} type="button" onClick={this.onCopy.bind(this, this.state.achievements[inx].id)}
                 value="Copy"></input>
        </td>
        <td className="pt-3-half">
          <div>  {this.state.achievements[inx].name} </div>
        </td>
        <td style={{"width": "25%"}} className="pt-3-half">
          <div>   {this.state.achievements[inx].event} </div>
        </td>
        <td style={{"width": "15%"}} className="text-right">
          <div>  {this.state.achievements[inx].condition} </div>
        </td>
        <td style={{"width": "15%"}} className="text-right">
          <div>  {this.state.achievements[inx].reward} </div>
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
                  Achievements
                </CardHeader>
                <CardBody>

                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">Id</th>
                      <th className="text-center">Name</th>
                      <th className="text-center">Event</th>
                      <th className="text-center">Condition</th>
                      <th className="text-center">Reward</th>
                      <th></th>
                    </tr>
                    </thead>
                    <tbody>
                    {roundsList}
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

export default Achievements;
/*        <YesNoModal ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
        </YesNoModal>*/
