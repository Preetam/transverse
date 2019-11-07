import React from 'react';
import ReactDOM from 'react-dom';
import App from './App';
const root = document.createElement('div');
document.title = 'Transverse';
root.style.cssText = 'min-height: 100%; display: flex; flex-direction: column;';
document.body.appendChild(root);
ReactDOM.render(<App />, root);
