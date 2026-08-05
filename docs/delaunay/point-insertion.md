# Point Insertion

The `engine/insertions.go` file handles inserting new points into the Delaunay triangulation. This is the core of the Bowyer-Watson algorithm — each point is located within the current mesh, the local structure is split, and affected edges are legalised.

## Algorithm Overview

```
For each new point P:
  1. Locate the triangle (or edge) containing P
  2. If P is inside a triangle → splitTriangle (creates 3 new triangles)
     If P is on an edge        → splitEdge     (creates 4 new triangles)
  3. Legalise every new edge via legaliseEdge (edge flips as needed)
  4. Update LastInsertedEdge for the next walk
```

## Function Reference

### `InsertPoint(m *Mesh, p Point) (*Mesh, error)`


The primary insertion function. Inserts a single point into the mesh and restores the Delaunay property.

**Step-by-step:**

1. **Add the point** to `m.Points` via `m.addPoint(p)`, receiving its `VertexID`.
2. **Walk** to find the triangle containing `p` via `walkToPoint`. Returns an edge of that triangle.
3. **Determine if P lies on an edge** of the containing triangle:
   - Iterates the 3 edges of the triangle.
   - For each edge AB, computes `orient(A, B, P)`.
   - If the signed area is `< Epsilon`, the point is collinear with that edge → treat as "on edge".
4. **Split:**
   - **On edge:** calls `splitEdge(m, edge, pID)` → returns 4 new edges.
   - **Inside triangle:** calls `splitTriangle(m, edge, pID)` → returns 3 new edges.
5. **Legalise** each returned edge via `legaliseEdge(m, edge, pID)`.
6. **Update** `m.LastInsertedEdge` to the first new edge (optimisation for subsequent walks).
7. Return the mesh and any error.

**Key insight:** The split step temporarily violates the Delaunay property (the new triangles may have points inside their circumcircles). The legalisation step restores it through edge flips.

---

### `walkToPoint(m *Mesh, p Point) (EdgeID, bool)`

Locates the triangle containing point `p` by walking through the mesh from a known starting edge. This is a **visibility walk** (also called an "oriented walk") — it follows the mesh topology using orientation tests to navigate toward the target.

**Algorithm:**

```
currentEdge = m.LastInsertedEdge

repeat up to len(Triangles) times:
    crossed = false

    for each of the 3 edges in the current triangle:
        // Skip tombstoned/boundary triangles
        if currentEdge.Triangle is None or tombstoned:
            jump to twin's next edge
            crossed = true
            break

        // Check if P is to the right of this edge
        if orient(A, B, P) < -Epsilon:    // P is on the right (outside)
            jump to twin's next edge      // Move to the adjacent triangle
            crossed = true
            break

        // P is to the left (inside) — advance to the next edge
        currentEdge = nextEdge

    // No jump means P is inside this triangle
    if not crossed:
        return currentEdge, true

return 0, false    // Failed to find
```

**Termination:** The walk is bounded by `len(Triangles)` iterations. In practice, for randomly distributed points, the expected number of steps is O(√n) (Devroye, Mücke & Zhu, 1998).

**Starting edge heuristic:** Each call starts from `m.LastInsertedEdge` — the edge created by the most recent insertion. This exploits spatial locality: consecutive insertions often land near each other, so the walk is typically very short (often O(1)).

---

### `orient(a, b, c Point) float64`

Computes the signed area of triangle (A, B, C) × 2 using the cross product:

```
orient = (B.x - A.x) * (C.y - A.y) - (B.y - A.y) * (C.x - A.x)
```

| Result | Meaning |
|---|---|
| `> 0` | C is to the **left** of AB (counter-clockwise) |
| `< 0` | C is to the **right** of AB (clockwise) |
| `≈ 0` | C is **collinear** with AB |

This is the fundamental primitive used throughout the algorithm — for point location (is P to the right of an edge?), edge detection (is P on an edge?), and the Delaunay circumcircle test.

---

### `splitTriangle(m *Mesh, AB EdgeID, p VertexID) [3]EdgeID`

Handles the case where a new point P lands **inside** a triangle. Destroys the original triangle and creates 3 new ones.

**Given triangle ABC with edges AB, BC, CA and new point P:**

1. **Tombstone** the original triangle.
2. **Create 3 new triangles:** ABP, BCP, CAP (each references one original edge).
3. **Create 3 new edge pairs:** BP↔PB, CP↔PC, AP↔PA.
4. **Rewire** the `Next` pointers and `Triangle` assignments for all 6 half-edges (3 original + 3 new).

**Before:**
```
        A
       / \
      /   \
     /     \
    /       \
   B ------- C
```

**After:**
```
               A
              /|\
             / | \
            /  |  \
           /   |   \
          /    |    \
         /     |     \
        /      P      \
       /     /   \     \
      /    /       \    \
     /   /           \   \
    /  /               \  \
   / /                   \ \
  B ----------------------- C
```

**Returns:** `[BP, CP, AP]` — the 3 new edges incident to P. These are the edges that need legalisation.

---

### `splitEdge(m *Mesh, AB EdgeID, p VertexID) [4]EdgeID`

**File:** `engine/insertions.go:134`

Handles the case where a new point P lands **on an existing edge** AB. This edge is shared by two adjacent triangles (ABC on one side, ABD on the other). Both triangles are destroyed and replaced by 4 new triangles.

**Given edge AB shared by triangles ABC and ABD, with new point P on AB:**

1. **Tombstone** both triangles (ABC via `m.HalfEdges[AB].Triangle`, ABD via `m.HalfEdges[BA].Triangle`).
2. **Create 4 new triangles:** APC, CPB, DPA, BPD.
3. **Create 3 new edge pairs:** PB↔BP, PC↔CP, DP↔PD.
4. **Repurpose existing edges:** AB becomes AP, BA becomes PA (their origins and `Next` pointers change).
5. **Rewire** all `Next` pointers and `Triangle` assignments.

**Before:**
```
   A---------C
   |\        |
   | \       |
   |  \      |
   |   \     |
   |    \    |
   |     \   |
   |      \  |
   |       \ |
   |        \|
   D---------B

```

**After:**
```
   A---------C
   |\       /|
   | \     / |
   |  \   /  |
   |   \ /   |
   |    P    |
   |   / \   |
   |  /   \  |
   | /     \ |
   |/       \|
   D---------B

```

**Returns:** `[AB, BP, CP, DP]` — the 4 new edges incident to P (using repurposed AB). These are the edges that need legalisation.

---

### Legalisation (called after split)

After `splitTriangle` or `splitEdge`, each returned edge is passed to `legaliseEdge(m, edge, pID)`. This recursively flips edges that violate the Delaunay (empty circumcircle) condition. See [legalisation.md](./legalisation.md) for details.

---

## Complexity

| Operation | Time | Notes |
|---|---|---|
| `walkToPoint` | O(√n) expected | For random point distributions; O(n) worst case |
| `splitTriangle` | O(1) | Constant work: 3 new triangles, 3 edge pairs |
| `splitEdge` | O(1) | Constant work: 4 new triangles, 3 edge pairs |
| `legaliseEdge` | O(k) per edge | k = number of flips (typically < 6, amortised constant) |
| **Per point** | **O(√n) expected** | Dominated by walk |
| **Total (n points)** | **O(n√n)** expected | Can be O(n log n) with better point location |

## Academic References

- **Bowyer, A.** (1981). "Computing Dirichlet tessellations." *The Computer Journal*, 24(2): 162–166.
- **Watson, D. F.** (1981). "Computing the n-dimensional Delaunay tessellation with application to Voronoi polytopes." *The Computer Journal*, 24(2): 167–172.
- **Devroye, L., Mücke, E. P. & Zhu, B.** (1998). "A note on point location in Delaunay triangulations of random points." *Algorithmica*, 22(4): 477–482.
