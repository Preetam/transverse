import styled from 'styled-components';

interface HeaderProps {
  flex?: number;
  alternate?: boolean;
}

export const Header = styled.header<HeaderProps>`
  padding: 10px;
  color: ${props =>
    props.alternate ? props.theme.alternateText : props.theme.text};
`;

export const H1 = styled.h1<HeaderProps>`
  padding: 10px;
  color: ${props =>
    props.alternate ? props.theme.alternateText : props.theme.text};
`;

export const H2 = styled.h2<HeaderProps>`
  padding: 10px;
  color: ${props =>
    props.alternate ? props.theme.alternateText : props.theme.text};
`;

export const H3 = styled.h3<HeaderProps>`
  padding: 10px;
  color: ${props =>
    props.alternate ? props.theme.alternateText : props.theme.text};
`;

export const H4 = styled.h4<HeaderProps>`
  padding: 10px;
  color: ${props =>
    props.alternate ? props.theme.alternateText : props.theme.text};
`;

export default Header;
