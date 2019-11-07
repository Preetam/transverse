import styled from 'styled-components';

interface BoxProps {
  flex?: number;
  alternate?: boolean;
  flexBasis?: string;
  flexGrow?: string;
}

const Box = styled.div<BoxProps>`
  display: flex;
  flex-direction: column;
  flex: ${props => props.flex ?? 'unset'};
  flex-grow: ${props => props.flexGrow ?? 'unset'};
  flex-basis: ${props => props.flexBasis ?? 'unset'};
  background-color: ${props =>
    props.alternate === true
      ? props.theme.alternateBackground
      : props.theme.background};
  color: ${props => props.theme.text};
`;

export default Box;
