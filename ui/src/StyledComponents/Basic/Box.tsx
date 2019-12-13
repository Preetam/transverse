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

interface BoxProps {
  alternate?: boolean;
}

const Box = styled.div<BoxProps & SpaceProps & FlexboxProps & LayoutProps>`
  ${space}
  ${layout}
  ${flexbox}
  display: flex;
  background-color: ${backgroundColor};
  color: ${color};
`;

export default Box;
