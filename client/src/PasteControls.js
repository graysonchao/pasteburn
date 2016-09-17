import React, { Component } from 'react';
import {
  Form,
  FormField,
  FormInput,
  Button,
  InputGroup
 } from 'elemental';
import './Paste.css';


class PasteControls extends Component {
  render() {
    return (
      <Form type="horizontal">
        <PasteKeyInput />
      </Form>
    )
  }
}

export default PasteControls;