const REFERENCE_WIDTH = 1920;
const REFERENCE_HEIGHT = 1080;

//everything on screen is laid out for the reference size and shrunk by this factor on smaller screens
export function uiScale(width: number, height: number): number {
  return Math.min(1, width / REFERENCE_WIDTH, height / REFERENCE_HEIGHT);
}
