const MINOR_PER_MAJOR = 100;

export const majorToMinor = (major: number): number => {
  if (!Number.isFinite(major)) {
    throw new Error('majorToMinor: value must be a finite number');
  }
  return Math.round(major * MINOR_PER_MAJOR);
};

export const minorToMajor = (minor: number): number => {
  if (!Number.isInteger(minor)) {
    throw new Error('minorToMajor: value must be an integer');
  }
  return minor / MINOR_PER_MAJOR;
};
