import styled from 'styled-components';

interface HeaderProps {
  flex?: number;
}

const Header = styled.header<HeaderProps>`
  padding: 10px;
  color: ${props => props.theme.text};
`;

export default Header;
