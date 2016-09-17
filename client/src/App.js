import React, { Component } from 'react';
import Pasteburn from './Pasteburn';
import './App.css';
import './styles.min.css'

class App extends Component {
  render() {
    return (
      <div className="App">
        <div style={{ margin: '5% 10%' }}>
          <Pasteburn />
        </div>
      </div>
    );
  }
}

export default App;
