import {WML} from "@wailsio/runtime";
import {EngineService} from "../bindings/github.com/ORBWARRIOR/PolyNav";

// Wire up data-wml-openURL links (logos + footer "Docs" link) once the DOM is ready.
WML.Enable();

const GRID_SIZE = 40; // 40px
const POINT_RADIUS = 7; // 7px

const sidebar = document.getElementById("sidebar");
const sidebarToggle = document.getElementById("sidebar-toggle");

const addObs = document.getElementById("add-obstacle");
const delObs = document.getElementById("delete-obstacle");
const addStart = document.getElementById("add-start");
const addEnd = document.getElementById("add-end");
const delStart = document.getElementById("delete-start");
const delEnd = document.getElementById("delete-end");
const triangulate = document.getElementById("triangulate");
const canvas = document.getElementById("canvas") as HTMLCanvasElement;
const ctx = canvas.getContext("2d")!;

if (!sidebar || !sidebarToggle || !addObs || !delObs || !addStart || !addEnd || !delStart || !delEnd || !triangulate || !canvas) {
    let msg = "Required DOM elements were not found.";
    EngineService.Log(msg)
    throw new Error(msg);
}

let activeTool: string | null;

const toolButtons = [addObs, delObs, addStart, delStart, addEnd, delEnd];

toolButtons.forEach((button) => {
    button.addEventListener("click", () => {
        const toolId = button.id;

        if (activeTool === toolId) {
            activeTool = null;
            button.classList.remove("selected");
        } else {
            const prev = activeTool ? document.getElementById(activeTool) : null;
            prev?.classList.remove("selected");
            activeTool = toolId;
            button.classList.add("selected");
        }
    });
});

sidebarToggle.addEventListener("click", () => {
    const isOpen = sidebar.classList.toggle("open");
    sidebarToggle.classList.toggle("open", isOpen);
})

triangulate.addEventListener("click", () => {
    EngineService.Log("triangulate clicked");
})

function resizeCanvas() {
    const dpr = window.devicePixelRatio || 1;
    canvas.width = window.innerWidth * dpr;
    canvas.height = window.innerHeight * dpr;
    ctx.scale(dpr, dpr);
    drawGraph();
}

window.addEventListener("resize", resizeCanvas);

function drawGraph() {
    const width = canvas.width;
    const height = canvas.height;
    const originX = width / 2
    const originY = height / 2

    ctx.clearRect(0, 0, width, height)
    ctx.strokeStyle = "rgba(255, 255, 255, 0.9)";
    ctx.lineWidth = 0.275;

    for (let x = originX % GRID_SIZE; x < width; x+=GRID_SIZE) {
        ctx.beginPath();
        ctx.moveTo(x, 0)
        ctx.lineTo(x, height)
        ctx.stroke();
    }
    
    for (let y = originY % GRID_SIZE; y < height; y+=GRID_SIZE) {
        ctx.beginPath();
        ctx.moveTo(0, y)
        ctx.lineTo(width, y)
        ctx.stroke();
    }

    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(originX, 0)
    ctx.lineTo(originX, height)
    ctx.stroke();

    ctx.beginPath();
    ctx.moveTo(0, originY)
    ctx.lineTo(width, originY)
    ctx.stroke();
}

resizeCanvas()
