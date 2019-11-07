import Box from '../Basic/Box';
import styled from 'styled-components';

export const EmptyBox = styled(Box)`
  align-items: center;
  justify-content: center;
`;

export const Row = styled(Box)`
  flex-direction: row;
`;

export const Column = styled(Box)`
  flex-direction: column;
`;

export default EmptyBox;
