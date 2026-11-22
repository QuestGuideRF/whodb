import { EdgeMarker, MarkerType, Node, Position, Viewport, XYPosition } from "reactflow";

function getNodeIntersection(node: Node, otherNode: Node, padding = 1) {

  const nodePos = node.positionAbsolute || { x: 0, y: 0 };
  const otherPos = otherNode.positionAbsolute || { x: 0, y: 0 };
  const nodeWidth = node.width || 100;
  const nodeHeight = node.height || 50;
  const otherWidth = otherNode.width || 100;
  const otherHeight = otherNode.height || 50;

  const nodeCenterX = nodePos.x + nodeWidth / 2;
  const nodeCenterY = nodePos.y + nodeHeight / 2;
  const otherCenterX = otherPos.x + otherWidth / 2;
  const otherCenterY = otherPos.y + otherHeight / 2;

  const dx = otherCenterX - nodeCenterX;
  const dy = otherCenterY - nodeCenterY;

  const scale = Math.max(
    Math.abs(dx) / ((nodeWidth + padding * 2) / 2),
    Math.abs(dy) / ((nodeHeight + padding * 2) / 2),
    1
  );

  return {
    x: nodeCenterX + (dx / scale) - Math.sign(dx) * padding,
    y: nodeCenterY + (dy / scale) - Math.sign(dy) * padding,
  };
}

function getEdgePosition(node: Node, intersectionPoint: { x: number; y: number }) {

    const nodePos = node.positionAbsolute || { x: 0, y: 0 };
    const nodeWidth = node.width || 100;
    const nodeHeight = node.height || 50;

    const px = intersectionPoint.x;
    const py = intersectionPoint.y;

    const isLeft = Math.abs(px - nodePos.x) <= 1;
    const isRight = Math.abs(px - (nodePos.x + nodeWidth)) <= 1;
    const isTop = Math.abs(py - nodePos.y) <= 1;
    const isBottom = Math.abs(py - (nodePos.y + nodeHeight)) <= 1;

    if (isLeft) return Position.Left;
    if (isRight) return Position.Right;
    if (isTop) return Position.Top;
    if (isBottom) return Position.Bottom;

  return Position.Top;
}

export function getEdgeParams(source: Node, target: Node, padding = 10) {
  const sourceIntersectionPoint = getNodeIntersection(source, target);
  const targetIntersectionPoint = getNodeIntersection(target, source);

  const sourcePos = getEdgePosition(source, sourceIntersectionPoint);
  const targetPos = getEdgePosition(target, targetIntersectionPoint);

  const dx = targetIntersectionPoint.x - sourceIntersectionPoint.x;
  const dy = targetIntersectionPoint.y - sourceIntersectionPoint.y;
  const length = Math.sqrt(dx * dx + dy * dy);

  const paddingX = length > 0 ? (dx / length) * padding : 0;
  const paddingY = length > 0 ? (dy / length) * padding : 0;

  return {
    sx: sourceIntersectionPoint.x + paddingX,
    sy: sourceIntersectionPoint.y + paddingY,
    tx: targetIntersectionPoint.x - paddingX,
    ty: targetIntersectionPoint.y - paddingY,
    sourcePos,
    targetPos,
  };
}

export function createNode(node: Pick<Node, "id" | "type" | "data"> & Partial<Node>, position?: XYPosition,  viewPort?: Viewport): Node {
  return {
      ...node,
      position: {
          x: position?.x != null ? position.x - 100 : (viewPort?.x ?? 100) / 2,
          y: position?.y != null ? position.y : (viewPort?.y ?? 100) / 2,
      },
      className: "group-[.laying-out]:transition-all group-[.laying-out]:duration-300 -z-5",
  }
}

export const MARKER_END: EdgeMarker = {
  type: MarkerType.ArrowClosed,
  width: 30,
  height: 30,
  orient: 'auto-start-reverse',
};

export function createEdge(source: string, target: string, type = "floatingGraphEdge") {
  return {
      id: `${source}->${target}`,
      type,
      source,
      target,
      markerEnd: MARKER_END,
      className: "transition-all duration-300 group-[.laying-out]:opacity-0",
  };
}