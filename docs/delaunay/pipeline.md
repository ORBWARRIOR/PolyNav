# Overall Pipeline

The `engine/delaunay.go` file orchestrates the full Delaunay triangulation from raw input points to a clean mesh. It follows the **Sloan incremental insertion algorithm** (Sloan, 1987), which combines Bowyer-Watson point insertion with Lawson edge-flip legalisation.

## Algorithm Overview

```
Input: []Point
  1. Normalise points to [0, 1] range
  2. Deduplicate coincident points
  3. Create super triangle enclosing all points
  4. Insert each point one at a time
  5. Legalise edges after each insertion
  6. Compact: remove super triangle artifacts
  7. Denormalise back to original coordinates
Output: *Mesh
```

## Function Reference

### `Triangulate(points []Point) (*Mesh, error)`


The top-level entry point. Runs the full pipeline:

1. **Guard:** returns an error if fewer than 3 points are provided.
2. **Normalise** the input points to the unit square `[0, 1]²` to improve floating-point stability.
3. **Deduplicate** points that are within `Epsilon` (1e-9) of each other.
4. **Guard:** returns an error if fewer than 3 unique points remain after deduplication.
5. **Build** a super triangle that encloses all normalised points via `NewMeshWithSuperTriangle`.
6. **Insert** each unique point sequentially via `InsertPoint`, which internally handles point location, triangle/edge splitting, and edge legalisation.
7. **Compact** the mesh to remove tombstoned triangles and half-edges created by the super triangle.
8. **Denormalise** the mesh points back to the original coordinate system.
9. Return the final `*Mesh`.

**Pseudocode:**
```
normalised, scale, minX, minY := normalisePoints(points)
uniq := deduplicatePoints(normalised)
mesh := NewMeshWithSuperTriangle(len(uniq))

for each point in uniq:
    InsertPoint(mesh, point)

compact(mesh)
mesh.Points = denormalisePoints(mesh.Points, scale, minX, minY)
return mesh
```

**References:**
- Sloan, S. W. (1987). "A fast algorithm for constructing Delaunay triangulations in the plane." *Advances in Engineering Software*, 9(1): 34–55.

---

### `NewMeshWithSuperTriangle(numOfPoints int) (*Mesh, error)`

Creates an initial mesh containing a single triangle large enough to enclose all input points. All subsequent insertions happen within this triangle.

**How it works:**

1. Pre-allocates slices with capacity estimates based on `numOfPoints`:
   - Points: `numOfPoints + 3` (input points + 3 super triangle vertices)
   - HalfEdges: `numOfPoints * 6`
   - Triangles: `numOfPoints * 2`
2. Adds 3 super triangle vertices at fixed coordinates `(-100, -100)`, `(100, -100)`, `(0, 100)` — chosen to enclose the `[0, 1]²` normalised range.
3. Adds 3 internal half-edges forming a counter-clockwise triangle (edges 0, 1, 2).
4. Adds 3 external boundary half-edges (edges 3, 4, 5) as clockwise twins with `Triangle: NoneTriangle` to represent the outer hull.
5. Sets `LastInsertedEdge = 2` (the last internal edge) as the starting point for subsequent `walkToPoint` calls.

**Mesh structure after creation:**

```
Points:    [(-100,-100), (100,-100), (0,100)]
           Vertex 0      Vertex 1    Vertex 2

HalfEdges: [0: AB→1, 1: BC→2, 2: CA→0]      // Internal (Triangle 0)
           [3: BA→5, 4: CB→3, 5: AC→4]      // External boundary (Triangle -1)

Triangles: [{Edge: 0, Tombstoned: false}]      // The super triangle
```

**Why a super triangle?**
The incremental algorithm requires every new point to land inside an existing triangle. By starting with a triangle that contains all points, this invariant is trivially satisfied. The super triangle's vertices are removed during `compact`.

---

### `normalisePoints(points []Point) (normalised []Point, scale, minX, minY float64)`

Transforms all input points into the `[0, 1]²` unit square. This normalisation improves numerical stability by keeping coordinates small and avoids floating-point precision issues in the circumcircle calculations.

**How it works:**

1. Finds the axis-aligned bounding box of all points (`minX`, `minY`, `maxX`, `maxY`).
2. Computes `scale = max(maxX - minX, maxY - minY)` — the larger dimension of the bounding box. If the range is negligible (`<= Epsilon`), defaults to `1.0`.
3. Transforms each point: `(x - minX) / scale`, `(y - minY) / scale`.
4. Returns the normalised points along with the transformation parameters needed for denormalisation.

**After normalisation:** All points lie in `[0, 1]²`, with at least one point touching each axis.

---

### `deduplicatePoints(points []Point) []Point`

Removes coincident points that are within `Epsilon` (1e-9) of each other. Duplicate points would cause degenerate triangles with zero area.

**How it works:**

1. Sorts points lexicographically by X, then by Y (using `Epsilon` tolerance for comparisons).
2. Walks through the sorted slice, keeping only points that differ from the previous kept point by more than `Epsilon`.
3. Returns the deduplicated slice.

**Helper:**
```go
func isCoincident(p1, p2 Point) bool {
    return math.Abs(p1.X-p2.X) < Epsilon && math.Abs(p1.Y-p2.Y) < Epsilon
}
```

---

### `compact(m *Mesh)`

**File:** `engine/delaunay.go:153`

Removes all tombstoned triangles and their associated half-edges from the mesh, then remaps all internal pointers. This is the final step that strips away the super triangle and any intermediate triangles destroyed during insertion/flipping.

**How it works:**

1. **Build new half-edge slice:** Iterates all half-edges. Skips any whose `Triangle` is tombstoned. Maps old indices to new indices.
2. **Build new triangle slice:** Iterates all triangles. Skips tombstoned ones. Maps old indices to new indices.
3. **Remap half-edge pointers:** Updates `Next`, `Twin`, and `Triangle` fields using the index maps. If a pointer's target was removed, it becomes `NoneEdge`/`NoneTriangle`.
4. **Remap triangle pointers:** Updates each triangle's `Edge` field.
5. **Update `LastInsertedEdge`:** Remaps to the new index, or falls back to the midpoint of the edge slice.

**Why boundary edges are preserved:**
Half-edges belonging to tombstoned triangles are removed, but boundary half-edges (those with `Triangle: NoneTriangle`) that are part of surviving triangle fans are kept. This is critical because these boundary edges close each triangle's `Next` chain — removing them would break mesh traversal.

**Post-compact:** The mesh contains only valid Delaunay triangles with no references to the super triangle.

---

### `denormalisePoints(points []Point, scale, minX, minY float64) []Point`

Reverses the normalisation by applying the inverse affine transform: `x * scale + minX`, `y * scale + minY`. This restores points to their original coordinate system.

---

### `MeshStats(m *Mesh) Stats`

Computes summary statistics for a mesh:

| Field | Computation |
|---|---|
| `PointCount` | `len(m.Points)` |
| `TriangleCount` | `len(m.Triangles)` |
| `EdgeCount` | `len(m.HalfEdges) / 2` (each undirected edge = 2 half-edges) |
| `HullEdges` | Count of half-edges where `Triangle == NoneTriangle` |

---

## Complexity

| Step | Time | Notes |
|---|---|---|
| Normalise | O(n) | Single pass over all points |
| Deduplicate | O(n log n) | Sort + linear scan |
| Super triangle | O(1) | Fixed structure |
| Insert all points | O(n × k) | k = average legalisation depth (typically small constant) |
| Compact | O(E + T) | E = half-edges, T = triangles |
| Denormalise | O(n) | Single pass |
| **Total** | **O(n log n)** dominated by sort | In practice, O(n) for uniform random points |

## Academic Reference

The overall pipeline implements the algorithm described in:

> **Sloan, S. W.** (1987). "A fast algorithm for constructing Delaunay triangulations in the plane." *Advances in Engineering Software*, 9(1): 34–55.

Sloan's contribution was combining Watson's incremental insertion (Bowyer-Watson) with Lawson's edge-flip legalisation into a single practical pipeline, using a super triangle for initialisation and a stack-based approach for managing the legalisation queue.
