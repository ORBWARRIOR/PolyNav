# Point Location

Point location is the problem of finding which triangle in the current mesh contains a given query point. This implementation uses a **visibility walk** — traversing triangle-to-triangle via orientation tests until the enclosing triangle is found.

## Algorithm Overview

The strategy is the **"last inserted edge" walking heuristic**:

1. Start from `m.LastInsertedEdge` (the edge created by the most recent insertion).
2. At each triangle, test the query point against all 3 edges using the `orient` function.
3. If the point is to the **right** of an edge (outside the triangle), cross to the adjacent triangle via that edge's twin.
4. If the point is to the **left** of all 3 edges, it is inside the current triangle — return it.
5. Repeat until found or iteration limit is reached.

This exploits **spatial locality**: when points are inserted in a spatially coherent order, consecutive points tend to land in nearby triangles, making the walk very short (often O(1)).

## Function Reference

### `walkToPoint(m *Mesh, p Point) (EdgeID, bool)`

The core point-location routine. Returns an edge of the triangle containing `p`, or `false` if the walk fails.

**Parameters:**
- `m` — the current mesh (must already contain at least one triangle)
- `p` — the query point to locate

**Returns:**
- `EdgeID` — one of the 3 edges of the triangle containing `p`
- `bool` — `true` if found, `false` if the walk failed (point is outside the mesh)

**Detailed algorithm:**

```
currentEdge = m.LastInsertedEdge
maxIter = len(m.Triangles)

for iter = 0 to maxIter:
    crossed = false

    // Walk up to 3 edges in the current triangle
    for step = 0 to 2:
        currentHE = m.HalfEdges[currentEdge]

        // Guard: skip tombstoned or boundary triangles
        if currentHE.Triangle is NoneTriangle or tombstoned:
            // Jump to the neighbouring triangle via twin
            currentEdge = m.HalfEdges[currentHE.Twin].Next
            crossed = true
            break

        nextEdge = currentHE.Next
        a = currentHE.Origin
        b = m.HalfEdges[nextEdge].Origin

        // Test: is P to the right of edge (a → b)?
        if orient(A, B, P) < -Epsilon:
            // P is outside this edge — cross to the adjacent triangle
            twin = currentHE.Twin
            if twin is NoneEdge:
                return 0, false    // Hit the hull boundary

            // Jump to the next edge in the neighbouring triangle
            currentEdge = m.HalfEdges[twin].Next
            crossed = true
            break

        // P is to the left (or on) this edge — advance to the next edge
        currentEdge = nextEdge

    // No crossing means P is inside the current triangle
    if not crossed:
        return currentEdge, true

// Exhausted iterations — walk failed
return 0, false
```

**Termination guarantee:** The walk is bounded by `len(m.Triangles)` iterations. Each iteration either advances to a new triangle or terminates. Since the mesh is finite, the walk must terminate.

**Failure modes:**
- **Point outside hull:** If `twin == NoneEdge` (hull boundary), the point cannot be inside any triangle → return `false`.
- **Iteration limit:** If the walk exceeds `len(Triangles)` steps without finding the point, it returns `false`. This can happen with pathological point distributions or degenerate meshes.

---

### `orient(a, b, c Point) float64`

The geometric primitive that powers the walk. Computes the signed area of triangle (A, B, C) × 2:

```
orient = (B.x - A.x) × (C.y - A.y) − (B.y − A.y) × (C.x − A.x)
```

This is the 2D cross product of vectors AB and AC.

| Result | Geometric Meaning | Walk Action |
|---|---|---|
| `> Epsilon` | C is to the **left** of directed edge AB | C is inside the triangle (for this edge) |
| `< -Epsilon` | C is to the **right** of directed edge AB | C is outside — cross to the neighbour |
| `≈ 0` (within ±Epsilon) | C is **collinear** with AB | C lies on edge AB — treated as "on edge" |

**Epsilon tolerance:** The comparison uses `-Epsilon` (not `0`) to handle floating-point imprecision. Points within `1e-9` of the edge are considered on the edge, not to the right.

---

### `LastInsertedEdge` Heuristic

The mesh stores `LastInsertedEdge` — an `EdgeID` pointing to the first edge created by the most recent `InsertPoint` call. Each subsequent `walkToPoint` starts from this edge.

**Why this works:**
- When inserting points in a spatially coherent order (e.g., sequential along a path), consecutive points tend to land in nearby triangles.
- Starting the walk from the last inserted edge means the walk is typically very short — often 0–3 steps.
- For truly random insertion order, the expected walk length is O(√n) per point (Devroye, Mücke & Zhu, 1998).

After `compact`: The `LastInsertedEdge` is remapped to the new edge index. If the original edge was removed (tombstoned), it falls back to `len(newEdges) / 2` — a midpoint guess.

---

## Walk Topology

The walk follows these navigation rules at each step:

```
Current triangle has edges: AB, BC, CA (counter-clockwise)
Query point P

Test AB: orient(A, B, P)
  → If P is right of AB: cross to neighbour via AB.Twin → BC' (start of adjacent triangle)
  → If P is left of AB: advance to edge BC

Test BC: orient(B, C, P)
  → If P is right of BC: cross via BC.Twin → CA' (start of adjacent triangle)
  → If P is left of BC: advance to edge CA

Test CA: orient(C, A, P)
  → If P is right of CA: cross via CA.Twin → AB' (start of adjacent triangle)
  → If P is left of CA: P is inside triangle ABC → return
```

**Key property:** The walk never visits the same triangle twice (in a valid Delaunay mesh), because each step moves strictly toward the query point.

---

## Complexity

| Metric | Value | Notes |
|---|---|---|
| Per step | O(1) | One `orient` call per edge |
| Steps per walk (random points) | O(√n) expected | Devroye, Mücke & Zhu (1998) |
| Steps per walk (spatially coherent) | O(1) expected | With `LastInsertedEdge` heuristic |
| Steps per walk (worst case) | O(n) | Pathological point distributions |

## Academic References

The walking point location strategy was first described by:

- **Lawson, C. L.** (1977). "Software for C¹ Surface Interpolation." *Mathematical Software III*, J. R. Rice (ed.), Academic Press, pp. 161–194.

- **Green, P. J. & Sibson, R.** (1978). "Computing Dirichlet tessellations in the plane." *The Computer Journal*, 21(2): 168–173.

The formal analysis of the expected walking cost was proven by:

- **Devroye, L., Mücke, E. P. & Zhu, B.** (1998). "A note on point location in Delaunay triangulations of random points." *Algorithmica*, 22(4): 477–482.

- **Mücke, E. P., Saias, I. & Zhu, B.** (1999). "Fast randomized point location without preprocessing in two- and three-dimensional Delaunay triangulations." *Computational Geometry: Theory and Applications*, 12(1–2): 63–83.

A robust variant with guaranteed termination was presented by:

- **Brown, P. J. C. & Faigle, C. T.** (1999). "A robust efficient algorithm for point location in triangulations." *Technical Report, University of Cambridge, Computer Laboratory*, UCAM-CL-TR-728.
