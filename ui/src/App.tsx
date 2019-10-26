import React from 'react';
import { ThemeProvider } from 'styled-components';
import { BrowserRouter as Router, Switch, Route, Link } from 'react-router-dom';
import Box from './Components/Basic/Box';

import { ChronicaProFontStyles, dark, light } from './Styles';

export default function App() {
  let theme = light;
  if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)')) {
    // theme = dark;
  }
  console.log(light);
  return (
    <ThemeProvider theme={theme}>
      <>
        <ChronicaProFontStyles />
        <Box>Hello World</Box>
      </>
    </ThemeProvider>
  );
}
