# Mesh Representation

The mesh is stored using the **half-edge data structure**, a well-established representation for polygon meshes and triangulations. This structure enables O(1) traversal of adjacent triangles, edges, and vertices.

## Academic Origin

The half-edge data structure is a simplified variant of the **quad-edge data structure** introduced by:

> **Guibas, L. & Stolfi, J.** (1985). "Primitives for the manipulation of general subdivisions and the computation of Voronoi diagrams." *ACM Transactions on Graphics*, 4(2): 74–123.

The quad-edge stores four directed half-edges per undirected edge (two for each direction, plus their duals). PolyNav uses a simpler half-edge variant that stores only one directed half-edge per record, with explicit twin pointers - sufficient for planar triangulations.

---

## Type Definitions

All types are defined in `engine/types.go`.

### Identity Types

```go
type VertexID    int32    // Index into Mesh.Points
type EdgeID      int32    // Index into Mesh.HalfEdges
type TriangleID  int32    // Index into Mesh.Triangles
```

These are `int32` aliases that prevent accidentally mixing indices. Sentinel values indicate "none":

```go
const NoneVertex   VertexID   = -1
const NoneEdge     EdgeID     = -1
const NoneTriangle TriangleID = -1
```

---

### `Point`

```go
type Point struct {
    X, Y float64
}
```

A 2D coordinate. Points are stored in `Mesh.Points` and referenced by `VertexID` (their index in the slice).

---

### `HalfEdge`

```go
type HalfEdge struct {
    Origin        VertexID   // Vertex where this edge starts
    Twin          EdgeID     // The opposite-direction half-edge (NoneEdge if on boundary)
    Next          EdgeID     // The next half-edge in the same triangle (counter-clockwise)
    Triangle      TriangleID // The triangle this half-edge belongs to (NoneTriangle if boundary)
    IsConstrained bool       // Reserved for Constrained Delaunay Triangulation (CDT)
}
```

**Each half-edge represents a directed segment** from `Origin` to the `Origin` of its `Next` edge. Together, three half-edges with matching `Triangle` values and linked via `Next` form a triangle.

**Key relationships:**

| Field | Meaning |
|---|---|
| `Origin` | The vertex this edge starts at (the destination is `m.HalfEdges[Next].Origin`) |
| `Twin` | The half-edge going the opposite direction. If `NoneEdge`, this edge is on the convex hull |
| `Next` | The next half-edge in counter-clockwise order around the same triangle |
| `Triangle` | The triangle this half-edge belongs to. `NoneTriangle` means it faces outward (boundary edge) |

**Half-edge pair convention:**
For every undirected edge between vertices A and B, there are exactly two half-edges:
- Edge `e`: `Origin = A`, `Twin = f`, pointing toward B
- Edge `f`: `Origin = B`, `Twin = e`, pointing toward A

---

### `Triangle`

```go
type Triangle struct {
    Edge       EdgeID  // One of the three half-edges bounding this triangle
    Tombstoned bool    // If true, this triangle has been deleted
}
```

A triangle is identified by any one of its three half-edges. The other two are found by following `Next` pointers:

```
Triangle T has Edge = e0

e0 → e1 = m.HalfEdges[e0].Next
e1 → e2 = m.HalfEdges[e1].Next
e2 → e0 = m.HalfEdges[e2].Next    // Closes the loop

Vertices:
  A = m.HalfEdges[e0].Origin
  B = m.HalfEdges[e1].Origin
  C = m.HalfEdges[e2].Origin
```

**Tombstoning:** Rather than removing triangles from the middle of the slice (which would invalidate all indices), triangles are marked as deleted by setting `Tombstoned = true`. The `compact` function later removes them and remaps all indices.

---

### `Mesh`

```go
type Mesh struct {
    Points           []Point     // All vertices, indexed by VertexID
    HalfEdges        []HalfEdge  // All directed edges, indexed by EdgeID
    Triangles        []Triangle  // All triangles, indexed by TriangleID
    LastInsertedEdge EdgeID      // Starting edge for the next walkToPoint call
}
```

The mesh is the complete data structure. All operations modify it in place.

**Memory layout:**

```
Points:    [P0, P1, P2, ..., Pn]                  // n+1 vertices (including super triangle)
HalfEdges: [HE0, HE1, HE2, ..., HE_m]             // m directed half-edges (m/2 undirected edges)
Triangles: [T0, T1, T2, ..., T_k]                 // k triangles (including tombstoned ones)
```

**Capacity hints (from `NewMeshWithSuperTriangle`):**
- Points: `numOfPoints + 3` (3 super triangle vertices)
- HalfEdges: `numOfPoints * 6` (each insertion adds ~6 half-edges)
- Triangles: `numOfPoints * 2` (Euler's formula: each insertion creates ~2 triangles)

---

## Traversal Patterns

### Walking around a triangle

```
edge := m.Triangles[tID].Edge
for i := 0; i < 3; i++ {
    vertex := m.HalfEdges[edge].Origin
    // Process vertex...
    edge = m.HalfEdges[edge].Next
}
```

### Getting all vertices of a triangle

```go
func (m *Mesh) GetTriangleVertices(tID TriangleID) (VertexID, VertexID, VertexID, bool)
```

Returns the three vertex IDs of triangle `tID`, or `(0, 0, 0, false)` if the triangle is invalid.

---

## Boundary Representation

Boundary (convex hull) edges are represented by half-edges with `Twin = NoneEdge` or `Triangle = NoneTriangle`.

```
Internal edge:                Boundary edge:
  e₀  →  e₁                   e₀  →  e₁
  │       │                   │       │
  T=0    T=0                  T=0    T=None
  │       │                   │       │
  e₃ ←  e₂                   e₃ ←  e₂

  Twin(e₀) = e₃               Twin(e₀) = NoneEdge
  Twin(e₁) = e₂               Twin(e₁) = NoneEdge
```

**Key invariant:** Boundary half-edges always have `Triangle = NoneTriangle`. This is how `MeshStats` counts hull edges, and how `walkToPoint` detects when the walk has hit the mesh boundary.

---

## Index Invariants

After `compact`, the following invariants hold:

1. **No dangling references:** Every `Next`, `Twin`, and `Triangle` pointer in the mesh references a valid index (or `None*` sentinel).
2. **No tombstoned triangles:** All remaining triangles have `Tombstoned = false`.
3. **Closed loops:** For every triangle, following `Next` three times returns to the starting edge.
4. **Twin symmetry:** If half-edge `e` has `Twin = f` (and `f ≠ NoneEdge`), then `m.HalfEdges[f].Twin = e`.
5. **Boundary consistency:** Boundary edges have `Twin = NoneEdge` and `Triangle = NoneTriangle`.

---

## Comparison with Other Data Structures

| Structure | Edges | Triangle access | Memory | Used in |
|---|---|---|---|---|
| **Half-edge** (this implementation) | 2 | Via `Triangle` field + `Next` chain | Low | PolyNav |
| Quad-edge (Guibas & Stolfi) | 4 (with duals) | Via primal/dual pointers | Higher | CGAL, classic literature |
| Winged-edge | 2 + adjacency lists | Via explicit face pointers | Higher | Early CAD systems |
| Face-based | 3 per face | Direct | Simple but redundant | Teaching implementations |

The half-edge structure is the a good choice for Delaunay triangulation implementations because it provides O(1) access to all adjacency information while keeping memory usage low.
