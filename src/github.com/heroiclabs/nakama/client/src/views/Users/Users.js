import React, {Component} from 'react';
import {getStyle, hexToRgba} from '@coreui/coreui/dist/js/coreui-utilities'
import {dashboard} from '../../App';
import ReactPaginate from 'react-paginate';
import {
  Card,
  CardBody,
  CardHeader,
  Col,
  Row,
  Table,
} from 'reactstrap';
import {YesNoModal} from "../Dashboard";

const brandPrimary = getStyle('--primary');


class Users extends Component {

  deleteCategory(cid, cname) {
    this.setState({
      currentCatId: cid,
      currentCatName: cname,
    });
  }

  deleteUser(uid) {
    dashboard.deleteUser(this.state.userToDelete, (res) => {
      this.update(this.state.page);
      //console.log(res);
    });
  }

  addCategory() {
    dashboard.addCategory((res) => {
      this.update(0);
      console.log(res);
    });
  }

  update(page) {
    dashboard.getUsers(page, 12, this.state.filterEmail, this.state.filterId, (res) => {
      this.setState({page: page, totalPages: res.payload.Pages, users: res.payload.Users ? res.payload.Users : []});
    });
  }

  editUser(u) {
    dashboard.editingUserId = u.Id;
    window.location = '#edit_user';
  }

  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.onRadioBtnClick = this.onRadioBtnClick.bind(this);


    this.state = {
      totalPages: 0,
      page: 0,
      users: [],
      dropdownOpen: false,
      radioSelected: 2,
      filterEmail: "",
      filterId: "",
    };

    if (!dashboard.session) {
      dashboard.RedirectLogin();
    } else {
      this.update(0);
    }
  }

  changeFilterId(event) {
    this.setState({filterId: (event.target.value)});
  }

  changeFilterEmail(event) {
    this.setState({filterEmail: (event.target.value)});
  }

  filter() {
    this.update(this.state.page);
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

  onPageChange(e) {
    let page = e.selected;
    this.update(page);
  }

  setUserToDelete(uid) {
    this.setState({userToDelete: uid})
  }

  render() {
    const buttonStyle = {border: "0px"};

    const userList = this.state.users.map((u) => {
      return <tr key={u.Id}>
        <td style={{"width": "5%"}}>
          <input className="btn btn-primary" style={buttonStyle} type="button" onClick={this.onCopy.bind(this, u.Id)}
                 value="Copy"></input>
        </td>
        <td className="text-center">
          <div>{u.Username}</div>
        </td>
        <td className="text-center">
          <div>{u.Email}</div>
        </td>
        <td className="text-center">
          <div>{u.Facebook_id}</div>
        </td>
        <td className="text-center">
          <div>{u.Google_id}</div>
        </td>
        <td className="text-center">
          <div>{new Date(u.Create_Time).toLocaleString("en-US")}</div>
        </td>
        <td className="text-center">
          <div>{u.Role > 0 ? "Yes" : "No"}</div>
        </td>

        <td style={{"width": "5%"}}>
          <span onClick={this.editUser.bind(this, u)} type="button" className="btn btn-default btn-sm"><span
            className="cui-pencil h5">
          </span>
          </span>
        </td>
        <td style={{"width": "5%"}} className="text-right">
          <button type="button" onClick={this.setUserToDelete.bind(this, u.Id)} data-toggle="modal"
                  data-target="#yesnomodal" style={buttonStyle} className="btn btn-danger">Delete
          </button>
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
                  Users
                </CardHeader>
                <CardBody>
                  <form>
                    <div className="form-group">
                      <label>Filter ID:</label>
                      <input onChange={this.changeFilterId.bind(this)} value={this.state.filterId}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Filter email:</label>
                      <input onChange={this.changeFilterEmail.bind(this)}
                             value={this.state.filterEmail}
                             className="form-control" type="text">
                      </input>
                    </div>
                    <div className="form-group">
                      <input className="btn btn-primary" onClick={this.filter.bind(this)} style={buttonStyle}
                             type="button" value="Filter">
                      </input>
                    </div>
                  </form>
                  <br/>



                  <Table hover responsive className="table-outline mb-0 d-none d-sm-table">
                    <thead className="thead-light">
                    <tr style={buttonStyle}>
                      <th className="text-center">ID</th>
                      <th className="text-center">Username</th>
                      <th className="text-center">Email</th>
                      <th className="text-center">Facebook Id</th>
                      <th className="text-center">Google Id</th>
                      <th className="text-center">Create Time</th>
                      <th className="text-center">Admin</th>
                      <th className="text-center"></th>
                      <th className="text-center"></th>
                    </tr>
                    </thead>
                    <tbody>
                    {userList}

                    </tbody>
                  </Table>
                  <br></br>
                  <div>
                  <nav aria-label="Page navigation example">
                    <ReactPaginate breakLabel={<a role="button" className="page-link">...</a>}
                                   previousLinkClassName={"page-link"}
                                   nextLinkClassName={"page-link"}
                                   nextLabel={"Next"}
                                   previousLabel={"Previous"}
                                   previousClassName={"page-item"}
                                   nextClassName={"page-item"}
                                   breakClassName={"page-item"}
                                   pageCount={this.state.totalPages}
                                   marginPagesDisplayed={2}
                                   pageRangeDisplayed={5}
                                   initialPage={this.state.page}
                                   pageClassName={"page-item"}
                                   onPageChange={this.onPageChange.bind(this)}
                                   containerClassName={"pagination justify-content-end"}
                                   pageLinkClassName={"page-link"}
                                   subContainerClassName={"page-link"}
                                   activeClassName={"active"}/>
                  </nav>
                  </div>
                </CardBody>
              </Card>
            </Col>
          </Row>

        </div>
        <YesNoModal ondelete={this.deleteUser.bind(this)} catid={this.state.currentCatId}
                    catname={this.state.currentCatName}>
        </YesNoModal>
      </div>
    );
  }
}

export default Users;
