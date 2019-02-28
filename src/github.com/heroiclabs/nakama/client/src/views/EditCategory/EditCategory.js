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


class EditCategory extends Component {
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

  addWord() {
    for (let x of this.state.category.languages) {
      x.words.push("");
    }
    this.setState({category: this.state.category});
  }

  updateCategory() {
    dashboard.updateCategory(this.state.category, (cat) => {
      //dashboard.editingCategory = cat.payload;
      this.update(this.state.category.id);

      this.setState({showSaveNot: true});
      setTimeout(()=>{
        this.setState({showSaveNot: false});
      }, 2000)
    });
  }

  deleteWord(inx) {
    for (let x of this.state.category.languages) {
      x.words.splice(inx, 1);
    }

    this.setState({category: this.state.category});
  }

  addCategory() {
    dashboard.addCategory((res) => {
      this.update();
      console.log(res);
    });
  }

  update(id) {
    dashboard.getCategory(id, (res) => {
      dashboard.editingCategory = res;
      this.setState({
        category: res,
      });
    });

    /*dashboard.getCategories((res) => {
      this.setState({
        categories: res,
      });
    });*/
  }

  constructor(props) {
    super(props);

    this.toggle = this.toggle.bind(this);
    this.onRadioBtnClick = this.onRadioBtnClick.bind(this);

    if (!dashboard.session) {
      dashboard.RedirectLogin();
    }

    this.state = {
      category: dashboard.editingCategory,
      categories: [],
      dropdownOpen: false,
      radioSelected: 2,
      currentCat: null,
      currentCatId: "",
      currentCatName: "",
      showSaveNot: false,
    };

    if (dashboard.editingCategory.name == undefined) dashboard.editingCategory.name = "";
    if (dashboard.editingCategory.unlock_price == undefined) dashboard.editingCategory.unlock_price = 0;

    this.update(this.state.category.id);
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

  changePrice(event) {
    this.state.category.unlock_price = parseFloat(event.target.value);
    event.target.value = this.state.category.unlock_price.toString()
    this.setState({category: this.state.category});
  }

  changeDesc(event) {
    this.state.category.description = event.target.value;
    this.setState({category: this.state.category});
  }

  changeName(event) {
    this.state.category.name = event.target.value;
    this.setState({category: this.state.category});
  }

  setWord(lang, word, event) {
    if (this.state.category.languages[lang]) {
      if (this.state.category.languages[lang].words[word] != undefined) {
        this.state.category.languages[lang].words[word] = event.target.value;
      } else {
        let len = this.state.category.languages[lang].words.length;
        let toAdd = word - len + 1;
        for (let i = 0; i < toAdd; i++) {
          this.state.category.languages[lang].words.push("");
        }
      }
    }
    this.setState({category: this.state.category});

  }

  getWord(lang, word, event) {
    let w = "";
    if (this.state.category.languages) {
      if (this.state.category.languages[lang].words[word]) {
        w = this.state.category.languages[lang].words[word];
      }
    }

    return w
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

    window.$('#img-upload').attr('src', this.state.category.imageData);


    const wordsList = [];
    let len = 0;
    for (let l of this.state.category.languages) {
      len = Math.max(len, l.words.length)
    }
    let imageTooBig = this.state.fileSize > 120000 ?<div className="successfully-saved alert alert-danger" role="alert">
      <strong>Image is too big: {Math.round(this.state.fileSize / 1000.)}KB</strong>
    </div>:[];
    let savedNotification = this.state.showSaveNot?<div className="successfully-saved alert alert-info" role="alert">
      <strong>Category saved</strong>
    </div>:[];

    for (let w = 0; w < len; w++) {
      wordsList.push(<tr key={w}>
        <td className="pt-3-half">
          <input onChange={this.setWord.bind(this, 0, w)} value={this.state.category.languages[0].words[w]}
                 className="form-control" id="inputdefault" type="text">
          </input>
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          <div >
            {this.state.category.languages[0].points[w]} ({this.state.category.languages[0].awards[w]})
          </div>
        </td>
        <td className="pt-3-half">
          <input onChange={this.setWord.bind(this, 1, w)} value={this.state.category.languages[1].words[w]}
                 className="form-control" id="inputdefault" type="text">
          </input>
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          <div >
            {this.state.category.languages[1].points[w]} ({this.state.category.languages[1].awards[w]})
          </div>
        </td>
        <td className="pt-3-half">
          <input onChange={this.setWord.bind(this, 2, w)} value={this.state.category.languages[2].words[w]}
                 className="form-control" id="inputdefault" type="text">
          </input>
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          <div >
            {this.state.category.languages[2].points[w]} ({this.state.category.languages[2].awards[w]})
          </div>
        </td>
        <td className="pt-3-half">
          <input onChange={this.setWord.bind(this, 3, w)} value={this.state.category.languages[3].words[w]}
                 className="form-control" id="inputdefault" type="text">
          </input>
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          <div >
            {this.state.category.languages[3].points[w]} ({this.state.category.languages[3].awards[w]})
          </div>
        </td>
        <td className="pt-3-half">
          <input onChange={this.setWord.bind(this, 4, w)} value={this.state.category.languages[4].words[w]}
                 className="form-control" id="inputdefault" type="text">
          </input>
        </td>
        <td style={{"width": "5%"}} className="pt-3-half">
          <div >
            {this.state.category.languages[4].points[w]} ({this.state.category.languages[4].awards[w]})
          </div>
        </td>

        <td style={{"width": "5%"}} className="text-right">
          <button type="button" onClick={this.deleteWord.bind(this, w)} style={buttonStyle}
                  className="btn btn-danger">Delete
          </button>
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
                  Edit category
                </CardHeader>
                <CardBody>
                  <form>
                    <div className="form-group">
                      <label>Name:</label>
                      <input onChange={this.changeName.bind(this)} value={this.state.category.name}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Description:</label>
                      <input onChange={this.changeDesc.bind(this)} value={this.state.category.description}
                             className="form-control">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Unlock price:</label>
                      <input pattern="[0-9]*" onChange={this.changePrice.bind(this)}
                             value={this.state.category.unlock_price ? this.state.category.unlock_price : "0"}
                             className="form-control" type="number">
                      </input>
                    </div>
                    <div className="form-group">
                      <label>Category image (jpg, 413x370px, max: 120kB)</label>
                      <div className="input-group">
            <span className="input-group-btn">
                  <input onChange={this.handleUploadFile.bind(this)} type="file" accept="image/jpeg" id="imgInp"></input>
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
                      <th className="text-center">English</th>
                      <th className="text-center">Points (Awards)</th>
                      <th className="text-center">Spanish</th>
                      <th className="text-center">Points (Awards)</th>
                      <th className="text-center">Italian</th>
                      <th className="text-center">Points (Awards)</th>
                      <th className="text-center">French</th>
                      <th className="text-center">Points (Awards)</th>
                      <th className="text-center">Portuguese</th>
                      <th className="text-center">Points (Awards)</th>
                      <th></th>
                    </tr>
                    </thead>
                    <tbody>
                    {wordsList}
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
                        <input className="btn btn-primary" onClick={this.addWord.bind(this)} style={buttonStyle}
                               type="button" value="Add word"></input>
                      </td>
                    </tr>
                    </tbody>
                  </Table>

                </CardBody>

              </Card>
              <Card>
                <CardBody>
                  {savedNotification}
                  <input className="btn btn-primary" onClick={this.updateCategory.bind(this)} style={buttonStyle}
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

export default EditCategory;
/*        <YesNoModal ondelete={this.delete.bind(this)} catid={this.state.currentCatId} catname={this.state.currentCatName}>
        </YesNoModal>*/
