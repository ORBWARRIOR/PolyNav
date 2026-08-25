import { WML } from "@wailsio/runtime";
import { EngineService } from "../bindings/github.com/ORBWARRIOR/PolyNav";
import type { Mesh } from "../bindings/github.com/ORBWARRIOR/PolyNav/engine";

// Wire up data-wml-openURL links (logos + footer "Docs" link) once the DOM is ready.
WML.Enable();

const sidebar = document.getElementById("sidebar");
const sidebarToggle = document.getElementById("sidebar-toggle");

const addObs = document.getElementById("add-obstacle");
const delObs = document.getElementById("delete-obstacle");
const addStart = document.getElementById("add-start");
const addEnd = document.getElementById("add-end");
const delStart = document.getElementById("delete-start");
const delEnd = document.getElementById("delete-end");
const clear = document.getElementById("clear");
const triangulate = document.getElementById("triangulate");
const canvas = document.getElementById("canvas") as HTMLCanvasElement;
const ctx = canvas.getContext("2d")!;

if (
	!sidebar ||
	!sidebarToggle ||
	!addObs ||
	!delObs ||
	!addStart ||
	!addEnd ||
	!delStart ||
	!delEnd ||
	!clear ||
	!triangulate ||
	!canvas
) {
	let msg = "Required DOM elements were not found.";
	EngineService.Log(msg);
	throw new Error(msg);
}

let activeTool: string | null = null;
const toolButtons = [addObs, delObs, addStart, delStart, addEnd, delEnd];

toolButtons.forEach((button) => {
	button.addEventListener("click", () => {
		const toolId = button.id;

		if (activeTool === toolId) {
			activeTool = null;
			button.classList.remove("selected");
		} else {
			const prev = activeTool
				? document.getElementById(activeTool)
				: null;
			prev?.classList.remove("selected");
			activeTool = toolId;
			button.classList.add("selected");
		}
	});
});

sidebarToggle.addEventListener("click", () => {
	const isOpen = sidebar.classList.toggle("open");
	sidebarToggle.classList.toggle("open", isOpen);
});

const GRID_SIZE = 40; // 40px
const POINT_RADIUS = 5; // 7px

type Point = {
	X: number;
	Y: number;
	Type: "start" | "end" | "obs";
};
let obstaclesPoints: Point[] = [];
let voronoiPoints: Point[] = [];

function drawCanvas() {
	const width = canvas.width;
	const height = canvas.height;
	const originX = width / 2;
	const originY = height / 2;

	ctx.clearRect(0, 0, width, height);
	ctx.strokeStyle = "rgba(255, 255, 255, 0.9)";
	ctx.lineWidth = 0.275;

	for (let x = originX % GRID_SIZE; x < width; x += GRID_SIZE) {
		ctx.beginPath();
		ctx.moveTo(x, 0);
		ctx.lineTo(x, height);
		ctx.stroke();
	}

	for (let y = originY % GRID_SIZE; y < height; y += GRID_SIZE) {
		ctx.beginPath();
		ctx.moveTo(0, y);
		ctx.lineTo(width, y);
		ctx.stroke();
	}

	ctx.lineWidth = 1;
	ctx.beginPath();
	ctx.moveTo(originX, 0);
	ctx.lineTo(originX, height);
	ctx.stroke();

	ctx.beginPath();
	ctx.moveTo(0, originY);
	ctx.lineTo(width, originY);
	ctx.stroke();

	obstaclesPoints.forEach((pt: Point) => {
		ctx.beginPath();
		ctx.arc(pt.X, pt.Y, POINT_RADIUS, 0, Math.PI * 2);
		ctx.fillStyle = "#ff1919";
		ctx.strokeStyle = "#ff1919";
		ctx.fill();
		ctx.lineWidth = 1;
		ctx.stroke();
	});

	voronoiPoints.forEach((pt: Point) => {
		ctx.beginPath();
		ctx.arc(pt.X, pt.Y, POINT_RADIUS, 0, Math.PI * 2);
		if (pt.Type === "start") {
			ctx.fillStyle = "#39e9ff";
			ctx.strokeStyle = "#39e9ff";
		} else {
			ctx.fillStyle = "#2cfe36";
			ctx.strokeStyle = "#2cfe36";
		}
		ctx.fill();
		ctx.lineWidth = 1;
		ctx.stroke();
	});
}

function filterPoints(
	targetType: Point["Type"],
	mouseX: number,
	mouseY: number,
) {
	if (targetType === "obs") {
		obstaclesPoints = obstaclesPoints.filter((pt) => {
			const dx = pt.X - mouseX;
			const dy = pt.Y - mouseY;
			const d = Math.sqrt(dx * dx + dy * dy);
			const clickRadius = POINT_RADIUS * 1.5;
			return d > clickRadius;
		});
	} else if (targetType === "start") {
		voronoiPoints = voronoiPoints.filter((pt) => {
			if (pt.Type !== "start") return true;
			const dx = pt.X - mouseX;
			const dy = pt.Y - mouseY;
			const d = Math.sqrt(dx * dx + dy * dy);
			const clickRadius = POINT_RADIUS * 1.5;
			return d > clickRadius;
		});
	} else {
		voronoiPoints = voronoiPoints.filter((pt) => {
			if (pt.Type !== "end") return true;
			const dx = pt.X - mouseX;
			const dy = pt.Y - mouseY;
			const d = Math.sqrt(dx * dx + dy * dy);
			const clickRadius = POINT_RADIUS * 1.5;
			return d > clickRadius;
		});
	}
}

canvas.addEventListener("click", (e: MouseEvent) => {
	if (!activeTool) return;
	const mouseX = e.clientX;
	const mouseY = e.clientY;

	switch (activeTool) {
		case addObs.id:
			obstaclesPoints.push({ X: mouseX, Y: mouseY, Type: "obs" });
			break;
		case addStart.id:
			voronoiPoints.find((pt, i) => {
				if (pt.Type === "start") voronoiPoints.splice(i, 1);
			});
			voronoiPoints.push({ X: mouseX, Y: mouseY, Type: "start" });
			break;
		case addEnd.id:
			voronoiPoints.find((pt, i) => {
				if (pt.Type === "end") voronoiPoints.splice(i, 1);
			});
			voronoiPoints.push({ X: mouseX, Y: mouseY, Type: "end" }) - 1;
			break;
		case delObs.id:
			filterPoints("obs", mouseX, mouseY);
			break;
		case delStart.id:
			filterPoints("start", mouseX, mouseY);
			break;
		case delEnd.id:
			filterPoints("end", mouseX, mouseY);
			break;
	}
	drawCanvas();
});

clear.addEventListener("click", () => {
	obstaclesPoints = [];
	voronoiPoints = [];
	drawCanvas();
});

function drawMesh(m: Mesh) {
	if (!m.HalfEdges) return;
	m.HalfEdges.forEach((halfEdge, i) => {
		if (!m.Points || !m.HalfEdges) return;
		const origin = m.Points[halfEdge.Origin];
		const next = m.Points[m.HalfEdges[halfEdge.Next].Origin];
		ctx.strokeStyle = "rgb(131, 131, 131)";
		ctx.beginPath();
		ctx.moveTo(origin.X, origin.Y);
		ctx.lineTo(next.X, next.Y);
		ctx.stroke();
	});
}

triangulate.addEventListener("click", async () => {
	if (obstaclesPoints.length < 3) return;
	try {
		const mesh = await EngineService.Triangulate(obstaclesPoints);
		if (mesh) drawMesh(mesh);
	} catch (err: unknown) {
		EngineService.Log(`Failed to triangulate: ${err}`);
	}
});

function resizeCanvas() {
	const dpr = window.devicePixelRatio || 1;
	canvas.width = window.innerWidth * dpr;
	canvas.height = window.innerHeight * dpr;
	ctx.scale(dpr, dpr);
	drawCanvas();
}
window.addEventListener("resize", resizeCanvas);
resizeCanvas();
