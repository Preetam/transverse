const fontSizeMap = new Map<string, number>(
  Object.entries({
    xsmall: 0.6,
    small: 0.8,
    medium: 1,
    large: 1.2,
    xlarge: 1.5,
  })
);

const spacingMap = new Map<string, number>(
  Object.entries({
    xsmall: 0.2,
    small: 0.4,
    medium: 1,
    large: 1.2,
    xlarge: 1.6,
  })
);

const fontSize = (percent: number = 1) => `${1 * percent}rem`;
export const namedFontSize = (size: string = 'medium') =>
  fontSize(fontSizeMap.get(size));

const spacing = (percent: number = 1) => `${1 * percent}rem`;
export const namedSpacing = (size: string = 'medium') =>
  spacing(spacingMap.get(size));
