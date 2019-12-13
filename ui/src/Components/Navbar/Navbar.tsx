import React from 'react';
import { Row } from '../../StyledComponents/Variants/Box';
import { H2 } from '../../StyledComponents/Basic/Header';
import Translations from '../../Contexts/Translations';

interface NavbarProps {}

const Navbar = (props: NavbarProps) => {
  const translations = React.useContext(Translations);
  return (
    <Row alternate>
      <H2 alternate pl="10px">
        {translations.appName}
      </H2>
    </Row>
  );
};

export default Navbar;
