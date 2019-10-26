import styled from 'styled-components';

interface BoxProps {
  flex?: number;
}

const Box = styled.div<BoxProps>`
  display: flex;
  flex-direction: column;
  flex: ${props => props.flex || 1};
  background-color: ${props => props.theme.background};
  color: ${props => props.theme.text};
`;

export default Box;
