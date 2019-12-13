interface Props {
  alternate?: boolean;
  theme: Styles;
}

export const backgroundColor = ({
  alternate,
  theme: { alternateBackground, background },
}: Props) => (alternate === true ? alternateBackground : background);

export const color = ({ alternate, theme: { alternateText, text } }: Props) =>
  alternate === true ? alternateText : text;
