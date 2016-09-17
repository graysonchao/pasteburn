import React, { Component } from 'react';
import {
  Row,
  Alert
 } from 'elemental';
import PasteContents from './PasteContents'
import PasteControls from './PasteControls'

class Pasteburn extends Component {
  constructor() {
    super();

    this.state = {
      contents: 'contents'
    }
  }

  render() {
    return (
      <div id="app">
        <Row>
          <h1>Pasteburn!</h1>
        </Row>
        <PasteContents />
      </div>
    )
  }
}

export default Pasteburn;
