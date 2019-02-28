import React, { Component } from 'react';
import { getStyle, hexToRgba } from '@coreui/coreui/dist/js/coreui-utilities'
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

// Card Chart 1
const cardChartData1 = {
  labels: ['January', 'February', 'March', 'April', 'May', 'June', 'July'],
  datasets: [
    {
      label: 'My First dataset',
      backgroundColor: brandPrimary,
      borderColor: 'rgba(255,255,255,.55)',
      data: [65, 59, 84, 84, 51, 55, 40],
    },
  ],
};

export class YesNoModal extends Component {

  constructor(props) {
    super(props);
  }

  delete(id, ondelete) {
    ondelete(id);
  }

  render() {
    return <div id="yesnomodal" className="modal" tabIndex="-1" role="dialog">
      <div className="modal-dialog" role="document">
        <div className="modal-content">
          <div className="modal-header">
            <h5 className="modal-title">Delete <strong>{this.props.catname}</strong>?</h5>
            <button type="button" className="close" data-dismiss="modal" aria-label="Close">
              <span aria-hidden="true">&times;</span>
            </button>
          </div>
          <div className="modal-body">
            <p>{this.props.nottext}</p>
          </div>
          <div className="modal-footer">
            <button type="button" onClick={this.delete.bind(this, this.props.catid, this.props.ondelete)} className="btn btn-danger" data-dismiss="modal">Yes</button>
            <button type="button" className="btn btn-secondary" data-dismiss="modal">No</button>
          </div>
        </div>
      </div>
    </div>
  }
}

class Dashboard extends Component {
  deleteCategory(cid, cname) {
    this.setState({
      currentCatId: cid,
      currentCatName: cname,
    });
   /* dashboard.deleteCategory(cid, (res)=>{
      this.update();
      console.log(res);
    });*/

  }

  delete(cid) {
    dashboard.deleteCategory(cid, (res)=>{
      this.update();
      console.log(res);
    });
  }

  addCategory() {
    dashboard.addCategory((res)=>{
      this.update();
      console.log(res);
      dashboard.editingCategory = res.payload;
      window.location = '#edit_category';
    });
  }

  update() {
    dashboard.getCategories((res)=>{
      this.setState({
        categories: res,
      });
    });
  }

  editCategory(cat) {
    dashboard.editingCategory = cat;
    window.location = '#edit_category';
  }

  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.onRadioBtnClick = this.onRadioBtnClick.bind(this);

    if (!dashboard.session) {
      dashboard.RedirectLogin();
    } else {
      this.update();
    }

    this.state = {
      categories: [],
      dropdownOpen: false,
      radioSelected: 2,
      currentCat: null,
      currentCatId: "",
      currentCatName: "",
    };

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

  render() {
    const buttonStyle = {border: "0px"};

    const categoriesList= this.state.categories.map((category) => {
      return <tr key={category.id}>
          <td style={{"width": "5%"}} >
            <input className="btn btn-primary" style={buttonStyle} type="button" onClick={this.onCopy.bind(this, category.id)} value="Copy"></input>
          </td>
        <td style={{"width": "8%"}} className="text-center">
          <div>{category.name}</div>
        </td>
        <td style={{"width": "12%"}} className="text-center">
          <div>{category.description}</div>
        </td>
        <td >
          <div>{(category.langMap["english"] != null && category.langMap["english"].words != null)?category.langMap["english"].words.join(', '):""}</div>
        </td>
        <td style={{"width": "10%"}} className="text-center">
          <div>
            {category.unlock_price>0?<strong>{category.unlock_price}</strong>:'0'}</div>
        </td>
        <td style={{"width": "5%"}}>
          <span style={buttonStyle} onClick={this.editCategory.bind(this, category)} type="button" className="btn btn-default btn-sm">
            <span className="cui-pencil h5"></span>
          </span>
        </td>
        <td style={{"width": "5%"}} className="text-right">
          <button type="button" onClick={this.deleteCategory.bind(this, category.id, category.name)}  data-toggle="modal" data-target="#yesnomodal" style={buttonStyle} className="btn btn-danger">Delete</button>
        </td>

      </tr>;
    });
    return (
      <div>
      <div className="animated fadeIn">
        <Row>
          <Col>
            <Card>
              <CardHeader>
                Categories
              </CardHeader>
              <CardBody>

                <br />
                <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                  <thead className="thead-light">
                  <tr style={buttonStyle}>
                    <th  className="text-center">ID</th>
                    <th className="text-center">Name</th>
                    <th >Description</th>
                    <th>Words</th>
                    <th className="text-center">Unlock Price</th>
                    <th></th>
                    <th></th>
                  </tr>
                  </thead>
                  <tbody>
                  {categoriesList}
                  <tr className="invalid">
                    <td >
                    </td>
                    <td className="text-center">
                    </td>
                    <td >
                    </td>
                    <td className="text-center">
                    </td>
                    <td>
                    </td>
                    <td>
                    </td>
                    <td className="text-right">
                      <input className="btn btn-primary" onClick={this.addCategory.bind(this)} style={buttonStyle} type="button" value="Add"></input>
                    </td>
                  </tr>
                  </tbody>
                </Table>
              </CardBody>
            </Card>
          </Col>
        </Row>

      </div>
      <YesNoModal nottext={"Are you sure you want to delele category '" + this.state.currentCatName + "'?"} ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
      </YesNoModal>
      </div>
    );
  }
}

export default Dashboard;
