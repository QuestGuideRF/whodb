import { FC, useCallback } from 'react';
import { ConnectionLineComponentProps, EdgeProps, Node, Position, getBezierPath, getSimpleBezierPath, useStore } from 'reactflow';
import { getEdgeParams } from './utils';

export const GraphEdgeConnectionLine: FC<ConnectionLineComponentProps> = ({ toX, toY, fromPosition, toPosition, fromNode }) => {
  if (!fromNode) {
    return null;
  }

  const targetNode: Node = {
    id: 'connection-target',
    width: 1,
    height: 1,
    positionAbsolute: { x: toX, y: toY },
    data: {},
    position: { x: toX, y: toY, },
  };

  const { sx, sy } = getEdgeParams(fromNode, targetNode);
  const [edgePath] = getBezierPath({
    sourceX: sx,
    sourceY: sy,
    sourcePosition: fromPosition,
    targetPosition: toPosition,
    targetX: toX,
    targetY: toY
  });

  return (
      <g>
        <path
            fill="none"
            strokeWidth={1.5}
            className="stroke-neutral-600 dark:stroke-neutral-300 animated"
            d={edgePath}
        />
        <circle
            cx={toX}
            cy={toY}
            r={3}
            className="fill-white dark:fill-white/10 stroke-neutral-600 dark:stroke-neutral-300"
            strokeWidth={1.5}
        />
      </g>
  );
}

export const FloatingGraphEdge: FC<EdgeProps> = ({ id, source, target, sourceHandleId, targetHandleId, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, markerEnd, style }) => {
  const sourceNode = useStore(useCallback((store) => store.nodeInternals.get(source), [source]));
  const targetNode = useStore(useCallback((store) => store.nodeInternals.get(target), [target]));

  if (!sourceNode || !targetNode) {
    return null;
  }

  // If handles are specified and we have coordinates from React Flow, use them

  let edgePath: string;

  if (sourceHandleId && targetHandleId && sourceX !== undefined && sourceY !== undefined && targetX !== undefined && targetY !== undefined) {

    [edgePath] = getBezierPath({
      sourceX,
      sourceY,
      sourcePosition: sourcePosition || Position.Right,
      targetPosition: targetPosition || Position.Left,
      targetX,
      targetY,
    });
  } else {

    const { sx, sy, tx, ty, sourcePos, targetPos } = getEdgeParams(sourceNode, targetNode);
    [edgePath] = getBezierPath({
      sourceX: sx,
      sourceY: sy,
      sourcePosition: sourcePos,
      targetPosition: targetPos,
      targetX: tx,
      targetY: ty,
    });
  }

  return (
    <path
      id={id}
      className="react-flow__edge-path"
      d={edgePath}
      markerEnd={markerEnd}
      style={style}
      data-edge-source={source}
      data-edge-target={target}
    />
  );
}