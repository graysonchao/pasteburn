import React, { Component } from 'react';
import {
   FormInput,
   FormField,
   Form,
   InputGroup,
   Button,
   Row,
   Card,
   Alert,
} from 'elemental';
import { isUUID } from 'validator';
import './Paste.css';

class PasteContents extends Component {
  constructor() {
    super();
    this.state = {
      key: "",
      id: "",
      body: ""
    };
    this.handleLoad = this.handleLoad.bind(this);
    this.handleSave = this.handleSave.bind(this);
    this.handleIDChange = this.handleIDChange.bind(this);
    this.handleKeyChange = this.handleKeyChange.bind(this);
    this.handleContentsChange = this.handleContentsChange.bind(this);
  }

  validate(uuid) {
    return isUUID(uuid, '4');
  }

  hasAlert() {
    return this.state && this.state.hasOwnProperty("alert") && this.state.alert.length > 0;
  }

  setAlert(message) {
    this.setState({alert: message});
    setTimeout(this.clearAlert.bind(this), 7000);
  }

  clearAlert() {
    this.setState({alert: ""});
  }

  handleIDChange(event) {
    this.setState({id: event.target.value});
  }

  handleKeyChange(event) {
    this.setState({key: event.target.value});
  }

  handleLoad() {
    if (this.state.key.length !== 32) {
      this.setAlert("Invalid key");
      return false;
    } else {
      fetch("http://127.0.0.1:8000/api/text/view?key=" + this.state.key + "&id=" + this.state.id, { method: "GET" })
      .then(function (res) {
        return res.json();
      })
      .then(function (doc) {
        this.setState(doc);
        console.log(doc);
        console.log(this.state);
      }.bind(this))
    }
  }

  handleSave() {
    if (this.state.key.length !== 32) {
      this.setAlert("Encryption key must be 32 characters long.");
      return false; 
    }
    fetch("http://127.0.0.1:8000/api/text/create", {
      method: 'POST',
      body: JSON.stringify({
        key: this.state.key,
        body: this.state.body
      })
    })
    .then(function (res) {
      return res.json();
    })
    .then(function (doc) {
      this.setState(doc);
    }.bind(this))
  }

  handleContentsChange(event) {
    this.setState({body: event.target.value});
  }

  render() {
    return (
      <Row>
        <Card>
          { this.hasAlert() ? <Alert type="danger">{this.state.alert}</Alert> : null }
          <Form type="horizontal">
            <InputGroup>
              <InputGroup.Section grow>
                <FormField>
                  <FormInput
                    type="text" 
                    placeholder="access key" 
                    id="documentKey"
                    onChange={this.handleKeyChange}
                    value={this.state.key}
                  />
                </FormField>
              </InputGroup.Section>
              <InputGroup.Section grow>
                <FormField>
                  <FormInput 
                    type="text" 
                    placeholder="document id" 
                    id="documentId"
                    onChange={this.handleIDChange}
                    value={this.state.id}
                  />
                </FormField>
              </InputGroup.Section>
              <InputGroup.Section>
                <Button type="primary" onClick={this.handleLoad}>Load</Button>
              </InputGroup.Section>
              <InputGroup.Section>
                <Button type="success" onClick={this.handleSave}>Save</Button>
              </InputGroup.Section>
            </InputGroup>
            <InputGroup>
              <FormField>
                <FormInput
                  placeholder="contents"
                  multiline
                  rows="24"
                  value={this.state.body}
                  id="documentContents"
                  onChange={this.handleContentsChange}
                /> 
              </FormField>
            </InputGroup>
          </Form>
        </Card>
      </Row>
    )
  }
}

export default PasteContents;
