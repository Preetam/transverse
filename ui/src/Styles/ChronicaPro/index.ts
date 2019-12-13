import { createGlobalStyle } from 'styled-components';

export default createGlobalStyle`
  
  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-ultralight.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-ultralight.woff') format('woff');
    font-weight: 200;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-ultralightit.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-ultralightit.woff') format('woff');
    font-weight: 200;
    font-style: italic;
  }

  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-book.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-book.woff') format('woff');
    font-weight: 300;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-bookit.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-bookit.woff') format('woff');
    font-weight: 300;
    font-style: italic;
  }

  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-regular.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-regular.woff') format('woff');
    font-weight: normal;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-regularit.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-regularit.woff') format('woff');
    font-weight: normal;
    font-style: italic;
  }

  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-bold.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-bold.woff') format('woff');
    font-weight: 700;
    font-style: normal;
  }

  @font-face {
    font-family: 'chronica pro';
    src: url('./src/Styles/ChronicaPro/chronicapro-boldit.woff2') format('woff2'),
        url('./src/Styles/ChronicaPro/chronicapro-boldit.woff') format('woff');
    font-weight: 700;
    font-style: italic;
  }

  body, html {
    font-family: 'chronica pro', sans-serif;
    min-height: 100%;
    height: 100%;
    padding: 0;
    margin: 0;
  }

`;
