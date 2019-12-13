import styled from 'styled-components';
import {
  space,
  layout,
  flexbox,
  SpaceProps,
  FlexboxProps,
  LayoutProps,
} from 'styled-system';

import { backgroundColor, color } from './Utils';

interface HeaderProps {
  alternate?: boolean;
  theme: Styles;
}

type Props = HeaderProps & FlexboxProps & LayoutProps & SpaceProps;

export const Header = styled.header<Props>`
  background-color: ${backgroundColor};
  color: ${color};
  ${space};
  ${layout};
`;

export const H1 = styled.h1<Props>`
  background-color: ${backgroundColor};
  color: ${color};
  ${space};
  ${layout};
`;

export const H2 = styled.h2<Props>`
  ${space};
  ${layout};
  background-color: ${backgroundColor};
  color: ${color};
`;

export const H3 = styled.h3<Props>`
  ${space};
  ${layout};
  background-color: ${backgroundColor};
  color: ${color};
`;

export const H4 = styled.h4<Props>`
  ${space};
  ${layout};
  background-color: ${backgroundColor};
  color: ${color};
`;

export default Header;
