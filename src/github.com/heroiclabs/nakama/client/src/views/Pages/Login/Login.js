import React, { Component } from 'react';
import { Button, Card, CardBody, CardGroup, Col, Container, Input, InputGroup, InputGroupAddon, InputGroupText, Row } from 'reactstrap';
import {dashboard} from '../../../App.js';

class Login extends Component {
  constructor (props) {
    super(props);

    this.state = {
      password: "",
      email: "",
      error: "",
      loggedIn: false,
    };
  }

  doLogin() {
    dashboard.Login(this.state.email, this.state.password, (err)=>{
      if (err) {
        this.setState({
          error: err,
        })
      }
    });
  }

  updatePassword (evt) {
    this.setState({
      password: evt.target.value
    });
  }

  updateEmail(evt) {
    this.setState({
      email: evt.target.value
    });
  }
  keyPress(e) {
    if(e.keyCode == 13){
       this.doLogin();
    }
  }

  render() {
    return (

      <div className="app flex-row align-items-center">

        <Container>
          <Row className="justify-content-center">
            <Col md="8">
              <CardGroup>
                <Card className="p-4">
                  <CardBody>
                    <h1>WordX Admin Login</h1>
                    <p className="text-muted">Sign In to your account</p>
                    {(this.state.error != "") &&
                    <div  className="alert alert-danger" role="alert">
                      {this.state.error}                    </div> }
                    <InputGroup className="mb-3">
                      <InputGroupAddon addonType="prepend">
                        <InputGroupText>
                          <i className="icon-user"></i>
                        </InputGroupText>
                      </InputGroupAddon>
                      <Input value={this.state.email} onChange={this.updateEmail.bind(this)} type="text" placeholder="Email" />
                    </InputGroup>
                    <InputGroup className="mb-4">
                      <InputGroupAddon addonType="prepend">
                        <InputGroupText>
                          <i className="icon-lock"></i>
                        </InputGroupText>
                      </InputGroupAddon>
                      <Input onKeyDown={this.keyPress.bind(this)}  onSubmit={this.doLogin.bind(this)} value={this.state.password} onChange={this.updatePassword.bind(this)} type="password" placeholder="Password" />
                    </InputGroup>
                    <Row>
                      <Col xs="6">
                        <Button color="primary" className="px-4" onClick={this.doLogin.bind(this)}>Login</Button>
                      </Col>
                      <Col xs="6" className="text-right">
                        {/*  <Button color="link" className="px-0">Forgot password?</Button> */}
                      </Col>
                    </Row>
                  </CardBody>
                </Card>
              </CardGroup>
            </Col>
          </Row>
        </Container>
      </div>
    );
  }
}

export default Login;
