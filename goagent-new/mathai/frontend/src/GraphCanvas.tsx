import { compile } from "mathjs";
import { useEffect, useRef, useState } from "react";
import type { GraphState } from "./types";

type Props = { graph: GraphState };

export function GraphCanvas({ graph }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [size, setSize] = useState({ width: 800, height: 600 });

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const observer = new ResizeObserver(([entry]) => setSize({ width: Math.max(320, entry.contentRect.width), height: Math.max(320, entry.contentRect.height) }));
    observer.observe(canvas);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ratio = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = Math.round(size.width * ratio);
    canvas.height = Math.round(size.height * ratio);
    const context = canvas.getContext("2d");
    if (!context) return;
    context.setTransform(ratio, 0, 0, ratio, 0, 0);
    drawGraph(context, size.width, size.height, graph);
  }, [graph, size]);

  return <canvas ref={canvasRef} className="graph-canvas" aria-label="函数图像" />;
}

function drawGraph(context: CanvasRenderingContext2D, width: number, height: number, graph: GraphState) {
  const { viewport, settings } = graph;
  const toX = (x: number) => (x - viewport.xMin) / (viewport.xMax - viewport.xMin) * width;
  const toY = (y: number) => height - (y - viewport.yMin) / (viewport.yMax - viewport.yMin) * height;
  context.clearRect(0, 0, width, height);
  context.fillStyle = "#ffffff";
  context.fillRect(0, 0, width, height);

  if (settings.showGrid) drawGrid(context, width, height, viewport, toX, toY);
  if (settings.showAxes) drawAxes(context, width, height, viewport, toX, toY);

  for (const equation of graph.equations.filter((item) => item.visible)) {
    try {
      const evaluator = compile(equation.expression);
      context.beginPath();
      context.strokeStyle = equation.color;
      context.lineWidth = equation.lineWidth;
      context.lineJoin = "round";
      let drawing = false;
      let previousY = 0;
      const samples = Math.max(100, Math.min(settings.sampleCount, 4000));
      for (let index = 0; index <= samples; index++) {
        const x = viewport.xMin + index / samples * (viewport.xMax - viewport.xMin);
        const value = evaluator.evaluate({ x });
        const y = typeof value === "number" ? value : Number.NaN;
        const screenX = toX(x), screenY = toY(y);
        const discontinuity = !Number.isFinite(y) || screenY < -height * 2 || screenY > height * 3 || (drawing && Math.abs(screenY - previousY) > height * 0.8);
        if (discontinuity) { drawing = false; continue; }
        if (!drawing) { context.moveTo(screenX, screenY); drawing = true; } else { context.lineTo(screenX, screenY); }
        previousY = screenY;
      }
      context.stroke();
    } catch {
      // The backend validates the expression language. A renderer error only skips this curve.
    }
  }
}

function drawGrid(context: CanvasRenderingContext2D, width: number, height: number, viewport: GraphState["viewport"], toX: (x: number) => number, toY: (y: number) => number) {
  context.strokeStyle = "#e8edf5";
  context.lineWidth = 1;
  context.fillStyle = "#8190a8";
  context.font = "11px system-ui";
  const xStep = niceStep(viewport.xMax - viewport.xMin);
  const yStep = niceStep(viewport.yMax - viewport.yMin);
  for (let x = Math.ceil(viewport.xMin / xStep) * xStep; x <= viewport.xMax; x += xStep) {
    const px = toX(x);
    context.beginPath(); context.moveTo(px, 0); context.lineTo(px, height); context.stroke();
    if (Math.abs(x) > xStep / 10) context.fillText(formatTick(x), px + 4, Math.min(height - 6, Math.max(14, toY(0) - 5)));
  }
  for (let y = Math.ceil(viewport.yMin / yStep) * yStep; y <= viewport.yMax; y += yStep) {
    const py = toY(y);
    context.beginPath(); context.moveTo(0, py); context.lineTo(width, py); context.stroke();
    if (Math.abs(y) > yStep / 10) context.fillText(formatTick(y), Math.min(width - 34, Math.max(5, toX(0) + 5)), py - 5);
  }
}

function drawAxes(context: CanvasRenderingContext2D, width: number, height: number, viewport: GraphState["viewport"], toX: (x: number) => number, toY: (y: number) => number) {
  context.strokeStyle = "#526077";
  context.lineWidth = 1.4;
  context.beginPath();
  if (viewport.yMin <= 0 && viewport.yMax >= 0) { context.moveTo(0, toY(0)); context.lineTo(width, toY(0)); }
  if (viewport.xMin <= 0 && viewport.xMax >= 0) { context.moveTo(toX(0), 0); context.lineTo(toX(0), height); }
  context.stroke();
}

function niceStep(range: number) {
  const rough = range / 10;
  const power = 10 ** Math.floor(Math.log10(rough));
  const fraction = rough / power;
  return (fraction < 2 ? 1 : fraction < 5 ? 2 : 5) * power;
}

function formatTick(value: number) { return Math.abs(value) >= 1000 ? value.toExponential(0) : Number(value.toFixed(4)).toString(); }
