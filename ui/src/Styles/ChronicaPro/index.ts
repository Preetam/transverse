import { createGlobalStyle } from 'styled-components';

export default createGlobalStyle`
  
  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-ultralight.woff2') format('woff2'),
        url('./chronicapro-ultralight.woff') format('woff');
    font-weight: 200;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-ultralightit.woff2') format('woff2'),
        url('./chronicapro-ultralightit.woff') format('woff');
    font-weight: 200;
    font-style: italic;
  }

  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-book.woff2') format('woff2'),
        url('./chronicapro-book.woff') format('woff');
    font-weight: 300;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-bookit.woff2') format('woff2'),
        url('./chronicapro-bookit.woff') format('woff');
    font-weight: 300;
    font-style: italic;
  }

  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-regular.woff2') format('woff2'),
        url('./chronicapro-regular.woff') format('woff');
    font-weight: normal;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-regularit.woff2') format('woff2'),
        url('./chronicapro-regularit.woff') format('woff');
    font-weight: normal;
    font-style: italic;
  }

  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-bold.woff2') format('woff2'),
        url('./chronicapro-bold.woff') format('woff');
    font-weight: 700;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica_pro';
    src: url('./chronicapro-boldit.woff2') format('woff2'),
        url('./chronicapro-boldit.woff') format('woff');
    font-weight: 700;
    font-style: italic;
  }

  body, html {
    font-family: chronica pro, Verdana, sans-serif;
    min-height: 100%;
    height: 100%;
    padding: 0;
    margin: 0;
  }

`;
