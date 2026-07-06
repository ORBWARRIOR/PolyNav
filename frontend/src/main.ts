import { Events } from "@wailsio/runtime";
import { EngineService } from "../bindings/github.com/ORBWARRIOR/PolyNav";
import type { Mesh, Point } from "../bindings/github.com/ORBWARRIOR/PolyNav/engine/models";

// ---- DOM refs -----------------------------------------------------------
const canvas = document.getElementById("mesh") as HTMLCanvasElement;
const generateBtn = document.getElementById("generate") as HTMLButtonElement;
const presetSelect = document.getElementById("preset") as HTMLSelectElement;
const statsEl = document.getElementById("stats") as HTMLSpanElement;
const timeEl = document.getElementById("time") as HTMLSpanElement;

// ---- State --------------------------------------------------------------
let mesh: Mesh | null = null;
let rawPoints: Point[] = [];
let dpr = devicePixelRatio || 1;

// ---- Point generators ---------------------------------------------------
function randomPoints(n: number): Point[] {
    const pts: Point[] = [];
    for (let i = 0; i < n; i++) {
        pts.push({ X: Math.random(), Y: Math.random() });
    }
    return pts;
}

function circlePoints(n: number): Point[] {
    const pts: Point[] = [];
    for (let i = 0; i < n; i++) {
        const a = (i / n) * Math.PI * 2;
        const r = 0.35 + Math.random() * 0.15;
        pts.push({ X: 0.5 + Math.cos(a) * r, Y: 0.5 + Math.sin(a) * r });
    }
    // sprinkle a few interior points
    for (let i = 0; i < Math.max(2, Math.floor(n / 4)); i++) {
        pts.push({ X: 0.5 + (Math.random() - 0.5) * 0.3, Y: 0.5 + (Math.random() - 0.5) * 0.3 });
    }
    return pts;
}

function gridPoints(n: number): Point[] {
    const side = Math.round(Math.sqrt(n));
    const pts: Point[] = [];
    for (let r = 0; r < side; r++) {
        for (let c = 0; c < side; c++) {
            pts.push({
                X: (c + 0.5 + (Math.random() - 0.5) * 0.3) / side,
                Y: (r + 0.5 + (Math.random() - 0.5) * 0.3) / side,
            });
        }
    }
    return pts;
}

const generators: Record<string, (n: number) => Point[]> = {
    "random-6": () => randomPoints(6),
    "random-12": () => randomPoints(12),
    "random-24": () => randomPoints(24),
    "circle-12": () => circlePoints(12),
    "grid-16": () => gridPoints(16),
};

// ---- Canvas sizing ------------------------------------------------------
function resizeCanvas(): void {
    const rect = canvas.getBoundingClientRect();
    dpr = devicePixelRatio || 1;
    canvas.width = rect.width * dpr;
    canvas.height = rect.height * dpr;
    const ctx = canvas.getContext("2d")!;
    ctx.scale(dpr, dpr);
    render();
}

const ro = new ResizeObserver(resizeCanvas);
ro.observe(canvas.parentElement!);

// ---- Rendering ----------------------------------------------------------
const PAD = 48;
const POINT_RADIUS = 3.5;
const HULL_RADIUS = 4.5;

function render(): void {
    const ctx = canvas.getContext("2d")!;
    const w = canvas.width / dpr;
    const h = canvas.height / dpr;

    // clear
    ctx.clearRect(0, 0, w, h);

    if (!mesh || !mesh.Points || mesh.Points.length === 0) {
        drawEmptyState(ctx, w, h);
        return;
    }

    const pts = mesh.Points;
    const vw = w - PAD * 2;
    const vh = h - PAD * 2;

    function toCanvas(p: Point): [number, number] {
        return [PAD + p.X * vw, PAD + p.Y * vh];
    }

    // ---- edges ----
    ctx.strokeStyle = "rgba(255,255,255,0.10)";
    ctx.lineWidth = 1;
    if (mesh.Edges) {
        for (const e of mesh.Edges) {
            const [x1, y1] = toCanvas(pts[e.A]);
            const [x2, y2] = toCanvas(pts[e.B]);
            ctx.beginPath();
            ctx.moveTo(x1, y1);
            ctx.lineTo(x2, y2);
            ctx.stroke();
        }
    }

    // ---- hull edges ----
    if (mesh.Boundary) {
        ctx.strokeStyle = "rgba(255,77,77,0.45)";
        ctx.lineWidth = 1.5;
        for (const e of mesh.Boundary) {
            const [x1, y1] = toCanvas(pts[e.A]);
            const [x2, y2] = toCanvas(pts[e.B]);
            ctx.beginPath();
            ctx.moveTo(x1, y1);
            ctx.lineTo(x2, y2);
            ctx.stroke();
        }
    }

    // ---- points ----
    for (const p of pts) {
        const [x, y] = toCanvas(p);
        // outer glow
        const glow = ctx.createRadialGradient(x, y, 0, x, y, POINT_RADIUS * 3);
        glow.addColorStop(0, "rgba(255,77,77,0.25)");
        glow.addColorStop(1, "rgba(255,77,77,0)");
        ctx.fillStyle = glow;
        ctx.beginPath();
        ctx.arc(x, y, POINT_RADIUS * 3, 0, Math.PI * 2);
        ctx.fill();

        // solid dot
        ctx.fillStyle = "#ff4d4d";
        ctx.beginPath();
        ctx.arc(x, y, POINT_RADIUS, 0, Math.PI * 2);
        ctx.fill();
    }
}

function drawEmptyState(ctx: CanvasRenderingContext2D, w: number, h: number): void {
    ctx.fillStyle = "rgba(255,255,255,0.06)";
    ctx.font = "13px -apple-system, sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText("Generate a triangulation to view the mesh", w / 2, h / 2);
}

// ---- Triangulate --------------------------------------------------------
async function runTriangulate(): Promise<void> {
    const preset = presetSelect.value;
    const gen = generators[preset];
    if (!gen) return;

    generateBtn.disabled = true;
    generateBtn.textContent = "Running…";

    try {
        rawPoints = gen();
        const raw = rawPoints;
        mesh = await EngineService.Triangulate(raw);
        render();
        updateStats();
    } catch (err) {
        console.error("Triangulate failed:", err);
        mesh = null;
        render();
    } finally {
        generateBtn.disabled = false;
        generateBtn.textContent = "Triangulate";
    }
}

function updateStats(): void {
    if (!mesh || !mesh.Points) {
        statsEl.textContent = "";
        return;
    }
    const n = mesh.Points.length;
    const tri = mesh.Indices ? mesh.Indices.length / 3 : 0;
    const edge = mesh.Edges ? mesh.Edges.length : 0;
    statsEl.textContent = `${n} pts · ${tri} tri · ${edge} edges`;
}

// ---- Event wiring -------------------------------------------------------
generateBtn.addEventListener("click", runTriangulate);
presetSelect.addEventListener("change", runTriangulate);

// ---- Time event from Go backend -----------------------------------------
Events.On("time", (event: { data: string }) => {
    const full = event.data;
    const compact = (full.match(/\d{1,2}:\d{2}:\d{2}/) || [full])[0];
    timeEl.textContent =
        window.matchMedia("(max-width: 640px)").matches ? compact : full;
});

// ---- Boot ---------------------------------------------------------------
document.getElementById("version")!.textContent = "v3.0.0-alpha2.107";
runTriangulate();
