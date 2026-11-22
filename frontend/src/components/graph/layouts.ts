import Dagre from '@dagrejs/dagre';
import { Edge, Node } from 'reactflow';

type ILayoutOption = {
    direction: "TB" | "LR";
}

const findConnectedComponents = (nodes: Node[], edges: Edge[]): Map<string, Set<string>> => {
    const nodeIds = new Set(nodes.map(n => n.id));
    const adjacency = new Map<string, Set<string>>();

    nodeIds.forEach(id => adjacency.set(id, new Set()));
    edges.forEach(edge => {
        adjacency.get(edge.source)?.add(edge.target);
        adjacency.get(edge.target)?.add(edge.source);
    });

    const visited = new Set<string>();
    const components = new Map<string, Set<string>>();
    let componentId = 0;

    const dfs = (nodeId: string, component: Set<string>) => {
        visited.add(nodeId);
        component.add(nodeId);
        adjacency.get(nodeId)?.forEach(neighbor => {
            if (!visited.has(neighbor)) {
                dfs(neighbor, component);
            }
        });
    };

    nodeIds.forEach(nodeId => {
        if (!visited.has(nodeId)) {
            const component = new Set<string>();
            dfs(nodeId, component);
            components.set(`component-${componentId++}`, component);
        }
    });

    return components;
};

const layoutComponent = (
    nodes: Node[],
    edges: Edge[],
    options: ILayoutOption
): { nodes: Node[]; width: number; height: number; minX: number; minY: number } => {
    const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}));

    g.setGraph({
        rankdir: options.direction,
        edgesep: 50,
        nodesep: 80,
        ranksep: 120,
        marginx: 40,
        marginy: 40,
        align: "UL",
        acyclicer: "greedy",
        ranker: "network-simplex",
    });

    edges.forEach((edge) => g.setEdge(edge.source, edge.target));

    nodes.forEach((node) => {

        const width = node.width || 400;
        const height = node.height || 200;

        const nodeWithDimensions = {
            ...node,
            width,
            height,
        };
        g.setNode(node.id, nodeWithDimensions as any);
    });

    Dagre.layout(g);

    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
    const layoutedNodes = nodes.map((node) => {
        const dagreNode = g.node(node.id);
        const x = dagreNode?.x || 0;
        const y = dagreNode?.y || 0;

        const width = node.width || 400;
        const height = node.height || 200;

        minX = Math.min(minX, x - width / 2);
        minY = Math.min(minY, y - height / 2);
        maxX = Math.max(maxX, x + width / 2);
        maxY = Math.max(maxY, y + height / 2);

        return {
            ...node,
            position: { x, y }
        };
    });

    return {
        nodes: layoutedNodes,
        width: maxX - minX,
        height: maxY - minY,
        minX,
        minY,
    };
};

const packComponents = (
    components: Array<{ nodes: Node[]; width: number; height: number; minX: number; minY: number }>
): Node[] => {
    if (components.length === 0) return [];
    if (components.length === 1) {

        const component = components[0];
        return component.nodes.map(node => {
            const width = node.width || 400;
            const height = node.height || 200;

            const centerX = node.position.x - component.minX;
            const centerY = node.position.y - component.minY;
            const topLeftX = centerX - width / 2;
            const topLeftY = centerY - height / 2;

            return {
                ...node,
                position: {
                    x: topLeftX,
                    y: topLeftY,
                }
            };
        });
    }

    const sorted = [...components].sort((a, b) => (b.width * b.height) - (a.width * a.height));

    const COMPONENT_SPACING = 150;

    const targetAspectRatio = 1.5;
    const cols = Math.max(1, Math.ceil(Math.sqrt(sorted.length * targetAspectRatio)));

    const grid: Array<Array<typeof sorted[0]>> = [];
    let rowIndex = 0;
    let colIndex = 0;

    sorted.forEach((component) => {
        if (colIndex >= cols) {
            rowIndex++;
            colIndex = 0;
        }

        if (!grid[rowIndex]) {
            grid[rowIndex] = [];
        }

        grid[rowIndex][colIndex] = component;
        colIndex++;
    });

    const rowHeights = grid.map(row =>
        Math.max(...row.map(comp => comp.height))
    );

    const rowYPositions = rowHeights.reduce((acc, _height, index) => {
        if (index === 0) {
            acc.push(0);
        } else {
            acc.push(acc[index - 1] + rowHeights[index - 1] + COMPONENT_SPACING);
        }
        return acc;
    }, [] as number[]);

    const allNodes: Node[] = [];

    grid.forEach((row, rIdx) => {
        let currentX = 0;
        const currentY = rowYPositions[rIdx];

        row.forEach((component) => {

            const offsetX = currentX - component.minX;
            const offsetY = currentY - component.minY;

            component.nodes.forEach(node => {
                const width = node.width || 400;
                const height = node.height || 200;

                const centerX = node.position.x + offsetX;
                const centerY = node.position.y + offsetY;
                const topLeftX = centerX - width / 2;
                const topLeftY = centerY - height / 2;

                allNodes.push({
                    ...node,
                    position: {
                        x: topLeftX,
                        y: topLeftY,
                    }
                });
            });

            currentX += component.width + COMPONENT_SPACING;
        });
    });

    return allNodes;
};

export const getDagreLayoutedElements = (nodes: Node[] = [], edges: Edge[] = [], options: ILayoutOption = {
    direction: "LR",
}) => {
    if (nodes.length === 0) {
        return { nodes: [], edges };
    }

    const componentMap = findConnectedComponents(nodes, edges);

    if (componentMap.size === 1) {
        const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}));

        g.setGraph({
            rankdir: options.direction,
            edgesep: 50,
            nodesep: 80,
            ranksep: 120,
            marginx: 50,
            marginy: 50,
            align: "UL",
            acyclicer: "greedy",
            ranker: "network-simplex",
        });

        edges.forEach((edge) => g.setEdge(edge.source, edge.target));

        nodes.forEach((node) => {

            const width = node.width || 400;
            const height = node.height || 200;

            const nodeWithDimensions = {
                ...node,
                width,
                height,
            };
            g.setNode(node.id, nodeWithDimensions as any);
        });

        Dagre.layout(g);

        return {
            nodes: nodes.map((node) => {
                const dagreNode = g.node(node.id);
                const width = node.width || 400;
                const height = node.height || 200;

                const centerX = dagreNode?.x || 0;
                const centerY = dagreNode?.y || 0;
                const topLeftX = centerX - width / 2;
                const topLeftY = centerY - height / 2;

                return {
                    ...node,
                    position: {
                        x: topLeftX,
                        y: topLeftY
                    }
                };
            }),
            edges,
        };
    }

    const layoutedComponents: Array<{ nodes: Node[]; width: number; height: number; minX: number; minY: number }> = [];

    componentMap.forEach((nodeIds) => {
        const componentNodes = nodes.filter(n => nodeIds.has(n.id));
        const componentEdges = edges.filter(e => nodeIds.has(e.source) && nodeIds.has(e.target));

        const layouted = layoutComponent(componentNodes, componentEdges, options);
        layoutedComponents.push(layouted);
    });

    const finalNodes = packComponents(layoutedComponents);

    return {
        nodes: finalNodes,
        edges,
    };
};