export interface LayoutNode {
  id: number;
  connections: number[];
}

//the area a node claims; connections attach at its anchor, which sits anchorY below the box centre
export interface LayoutBox extends Size {
  anchorY: number;
}

interface Size {
  width: number;
  height: number;
}

interface Point {
  x: number;
  y: number;
}

const SPREAD = 1.6; // frame area per node, relative to the node's own box
const ITERATIONS = 300;
const BOX_PADDING = 40;
const MAX_SEPARATION_PASSES = 200;
const OBSTRUCTION_PUSH = 2; // relative to the ideal edge length; weaker than this, dense maps keep hidden connections
const JITTER = 0.2; // seed noise relative to the ideal edge length, keeps symmetric maps from stalling

//Fruchterman-Reingold: every node repels every other, connections pull their ends together,
//and a connection running through a third node's box pushes that node off its path.
//Returns the anchor of each node, the whole layout centred on 0,0.
export function layoutLocations(
  nodes: LayoutNode[],
  box: LayoutBox,
  aspect: number,
): Map<number, Point> {
  const area = nodes.length * box.width * box.height * SPREAD;
  const frame = { width: Math.sqrt(area * aspect), height: area / Math.sqrt(area * aspect) };
  const idealLength = Math.sqrt(area / nodes.length);

  const positions = seedOnGrid(nodes.length, frame, idealLength, aspect);
  settle(positions, edgesOf(nodes), frame, box, idealLength);
  separateOverlaps(positions, box);
  centre(positions, box);

  return new Map(nodes.map((node, i) => [node.id, positions[i]]));
}

function seedOnGrid(count: number, frame: Size, idealLength: number, aspect: number): Point[] {
  const columns = Math.ceil(Math.sqrt(count * aspect));
  const rows = Math.ceil(count / columns);
  return Array.from({ length: count }, (_, i) => ({
    x: ((i % columns) + 0.5) * (frame.width / columns) - frame.width / 2 + jitter(i, idealLength),
    y:
      (Math.floor(i / columns) + 0.5) * (frame.height / rows) -
      frame.height / 2 +
      jitter(i + count, idealLength),
  }));
}

//deterministic, so a given map always lays out the same way
function jitter(seed: number, idealLength: number): number {
  const noise = Math.sin(seed + 1) * 10000;
  return (noise - Math.floor(noise) - 0.5) * idealLength * JITTER;
}

//every connection once, whichever side lists it
function edgesOf(nodes: LayoutNode[]): [number, number][] {
  const indexOf = new Map(nodes.map((node, i) => [node.id, i]));
  const edges = new Map<string, [number, number]>();
  nodes.forEach((node, from) => {
    for (const id of node.connections) {
      const to = indexOf.get(id);
      if (to === undefined || to === from) continue;
      edges.set(`${Math.min(from, to)}-${Math.max(from, to)}`, [from, to]);
    }
  });
  return [...edges.values()];
}

function settle(
  positions: Point[],
  edges: [number, number][],
  frame: Size,
  box: LayoutBox,
  idealLength: number,
): void {
  const limitX = Math.max(frame.width - box.width, 0) / 2; // keeps the boxes, not just their anchors, inside the frame
  const limitY = Math.max(frame.height - box.height, 0) / 2;
  const startTemperature = frame.width / 10;
  for (let iteration = 0; iteration < ITERATIONS; iteration++) {
    const temperature = startTemperature * (1 - iteration / ITERATIONS);
    const displacement = positions.map(() => ({ x: 0, y: 0 }));

    for (let a = 0; a < positions.length; a++) {
      for (let b = a + 1; b < positions.length; b++) {
        const { dx, dy, distance } = offset(positions[a], positions[b]);
        const push = idealLength ** 2 / distance / distance;
        displacement[a].x += dx * push;
        displacement[a].y += dy * push;
        displacement[b].x -= dx * push;
        displacement[b].y -= dy * push;
      }
    }
    for (const [a, b] of edges) {
      const { dx, dy, distance } = offset(positions[a], positions[b]);
      const pull = distance / idealLength;
      displacement[a].x -= dx * pull;
      displacement[a].y -= dy * pull;
      displacement[b].x += dx * pull;
      displacement[b].y += dy * pull;
    }
    for (const [a, b] of edges) {
      positions.forEach((position, c) => {
        if (c === a || c === b) return;
        const side = sideOfBlockedPath(positions[a], positions[b], position, box);
        if (!side) return;
        const push = OBSTRUCTION_PUSH * idealLength;
        displacement[c].x += side.x * push;
        displacement[c].y += side.y * push;
        displacement[a].x -= (side.x * push) / 2;
        displacement[a].y -= (side.y * push) / 2;
        displacement[b].x -= (side.x * push) / 2;
        displacement[b].y -= (side.y * push) / 2;
      });
    }

    positions.forEach((position, i) => {
      const length = Math.hypot(displacement[i].x, displacement[i].y);
      if (length === 0) return;
      const step = Math.min(length, temperature) / length;
      position.x = clamp(position.x + displacement[i].x * step, limitX);
      position.y = clamp(position.y + displacement[i].y * step, limitY);
    });
  }
}

function offset(from: Point, to: Point) {
  const dx = from.x - to.x;
  const dy = from.y - to.y;
  return { dx, dy, distance: Math.max(Math.hypot(dx, dy), 0.01) };
}

function clamp(value: number, limit: number): number {
  return Math.max(-limit, Math.min(limit, value));
}

//when the connection from..to runs behind the node's box: the unit direction, perpendicular to the connection,
//in which the node has to move to get out of its way (the side it already leans to)
function sideOfBlockedPath(from: Point, to: Point, node: Point, box: LayoutBox): Point | null {
  const centre = { x: node.x, y: node.y - box.anchorY };
  const halfWidth = (box.width + BOX_PADDING) / 2;
  const halfHeight = (box.height + BOX_PADDING) / 2;
  if (!segmentHitsBox(from, to, centre, halfWidth, halfHeight)) return null;

  const length = Math.max(Math.hypot(to.x - from.x, to.y - from.y), 0.01);
  const normal = { x: -(to.y - from.y) / length, y: (to.x - from.x) / length };
  const lean = (centre.x - from.x) * normal.x + (centre.y - from.y) * normal.y;
  const sign = lean >= 0 ? 1 : -1;
  return { x: normal.x * sign, y: normal.y * sign };
}

//slab test of the segment against the box around `centre`
function segmentHitsBox(
  from: Point,
  to: Point,
  centre: Point,
  halfWidth: number,
  halfHeight: number,
): boolean {
  let entry = 0;
  let exit = 1;
  const axes = [
    { start: from.x - centre.x, delta: to.x - from.x, half: halfWidth },
    { start: from.y - centre.y, delta: to.y - from.y, half: halfHeight },
  ];
  for (const { start, delta, half } of axes) {
    if (delta === 0) {
      if (Math.abs(start) > half) return false;
      continue;
    }
    const near = (-half - start) / delta;
    const far = (half - start) / delta;
    entry = Math.max(entry, Math.min(near, far));
    exit = Math.min(exit, Math.max(near, far));
  }
  return entry <= exit;
}

//the force pass treats nodes as points, so overlapping boxes are pushed apart along the axis of least penetration
function separateOverlaps(positions: Point[], box: LayoutBox): void {
  const minDistanceX = box.width + BOX_PADDING;
  const minDistanceY = box.height + BOX_PADDING;
  for (let pass = 0; pass < MAX_SEPARATION_PASSES; pass++) {
    let moved = false;
    for (let a = 0; a < positions.length; a++) {
      for (let b = a + 1; b < positions.length; b++) {
        const dx = positions[b].x - positions[a].x;
        const dy = positions[b].y - positions[a].y;
        const overlapX = minDistanceX - Math.abs(dx);
        const overlapY = minDistanceY - Math.abs(dy);
        if (overlapX <= 0 || overlapY <= 0) continue;
        moved = true;
        if (overlapX < overlapY) {
          const half = (overlapX / 2) * (dx >= 0 ? 1 : -1);
          positions[a].x -= half;
          positions[b].x += half;
        } else {
          const half = (overlapY / 2) * (dy >= 0 ? 1 : -1);
          positions[a].y -= half;
          positions[b].y += half;
        }
      }
    }
    if (!moved) return;
  }
}

//centres the boxes, not the anchors
function centre(positions: Point[], box: LayoutBox): void {
  const xs = positions.map((p) => p.x);
  const ys = positions.map((p) => p.y - box.anchorY);
  const middleX = (Math.min(...xs) + Math.max(...xs)) / 2;
  const middleY = (Math.min(...ys) + Math.max(...ys)) / 2;
  for (const p of positions) {
    p.x -= middleX;
    p.y -= middleY;
  }
}
