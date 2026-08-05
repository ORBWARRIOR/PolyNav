# Legalisation

After inserting a point and splitting the local triangle/edge, the new triangles may violate the **Delaunay condition** (the empty circumcircle property). The legalisation step restores it by recursively flipping non-Delaunay edges.

## The Delaunay Condition

An edge is **locally Delaunay** if the opposite vertex of the adjacent triangle does not lie inside the circumcircle of the triangle on the other side. If it does, the edge must be **flipped** — replaced by the other diagonal of the quadrilateral formed by the two triangles.

```
Before flip:                    After flip:
       B                              B
      /|\                            / \
     / | \                          /   \
    /  |t2\    circumcircle        / t2  \
   P   |   C  contains C?         P-------C
    \t1|  /   → flip edge AB       \ t1  /
     \ | /                          \   /
      \|/                            \ /
       A                              A

   t1 = △PAB, t2 = △ACB            t1 = △PCB, t2 = △PAC
   Edge AB is flipped to PC
```

## Function Reference

### `legaliseEdge(m *Mesh, edge EdgeID, p VertexID)`

Recursively legalises an edge by checking the Delaunay condition and flipping if necessary.

**Parameters:**
- `m` — the mesh
- `edge` — an edge to check (incident to the newly inserted point `p`)
- `p` — the newly inserted point's `VertexID`

**Algorithm:**

```
legaliseEdge(m, edge, p):
    // Guard: edge is invalid or out of bounds
    if edge is NoneEdge or edge > len(HalfEdges):
        return

    // Guard: edge's triangle is tombstoned
    tID = m.HalfEdges[edge].Triangle
    if tID is NoneTriangle or m.Triangles[tID].Tombstoned:
        return

    // Find the edge opposite to P in the current triangle
    oppEdge = getEdgeOppositeP(m, edge, p)

    // Guard: the opposite edge has no twin (hull boundary)
    twin = m.HalfEdges[oppEdge].Twin
    if twin is NoneEdge:
        return

    // Guard: the twin's triangle is tombstoned
    twinTID = m.HalfEdges[twin].Triangle
    if twinTID is NoneTriangle or m.Triangles[twinTID].Tombstoned:
        return

    // Find vertex C (the vertex opposite AB in the adjacent triangle)
    A = m.HalfEdges[oppEdge].Origin
    B = m.HalfEdges[twin].Origin
    C = find third vertex of twin's triangle

    if C is NoneVertex:
        return

    // Delaunay test: does C lie inside the circumcircle of (A, B, P)?
    AC = m.HalfEdges[twin].Next
    CB = m.HalfEdges[AC].Next

    if pointInCircumcircle(C, P, A, B):
        // Violation! Flip edge oppEdge
        flipEdge(m, oppEdge)

        // Recursively check the two new suspect edges
        legaliseEdge(m, AC, p)
        legaliseEdge(m, CB, p)
```

**Key insight:** After a flip, the two new edges adjacent to the flipped edge may now violate the Delaunay condition. These are the edges `AC` and `CB` from the adjacent triangle. The recursion continues until no more violations are found or a boundary is reached.

**Termination:** Each flip removes one Delaunay violation. Since the number of possible triangulations of a fixed point set is finite, the recursion must terminate. In practice, the average number of flips per insertion is < 6 (Guibas, Knuth & Sharir, 1990).

---

### `flipEdge(m *Mesh, AB EdgeID)`

Performs a single edge flip. Given edge AB shared by two triangles (PAB and CAB), replaces AB with CD (or equivalently, CP).

**Parameters:**
- `m` — the mesh
- `AB` — the edge to flip (must have a valid twin)

**Before flip:**
```
         A
        /|\
       / | \
      /  |  \
     P   |   C      ← edge AB (to be flipped)
      \  |  /
       \ | /
        \|/
         B
```

**After flip:**
```
         A
        / \
       /   \
      /     \
     P-------C
      \     /
       \   /
        \ /
         B
```

**Algorithm:**

1. **Gather edges:** Walk the half-edge rings of both triangles to find all 6 edges: AB, BP, PA (triangle 1) and BA, AC, CB (triangle 2).
2. **Extract vertices:** P from triangle 1, C from triangle 2.
3. **Tombstone** both original triangles (t1 and t2).
4. **Create 2 new triangles** (t3 and t4).
5. **Repurpose** AB to become edge PC (change origin from A to P, rewire `Next`).
6. **Repurpose** BA to become edge CP (change origin from C, rewire `Next`).
7. **Assign** all edges to their new triangles.

**Wire-up details:**

```
// Triangle t3: edges BP → AB(PC) → CB
m.HalfEdges[BP].Next = AB
m.HalfEdges[AB].Origin = P
m.HalfEdges[AB].Next = CB
m.HalfEdges[CB].Next = BP
// All three → Triangle t3

// Triangle t4: edges AC → BA(CP) → PA
m.HalfEdges[AC].Next = BA
m.HalfEdges[BA].Origin = C
m.HalfEdges[BA].Next = PA
m.HalfEdges[PA].Next = AC
// All three → Triangle t4
```

**Returns:** Nothing (modifies `m` in place).

---

### `getEdgeOppositeP(m *Mesh, seedEdge EdgeID, p VertexID) EdgeID`

Given an edge incident to point `p`, returns the edge opposite to `p` in the same triangle. This is the edge that needs to be tested for the Delaunay condition.

**Logic:**

```
if seedEdge.Origin == p:
    return seedEdge.Next          // The next edge starts where seedEdge ends — opposite to p
else if seedEdge.Next.Origin == p:
    return seedEdge.Next.Next     // The edge after that is opposite to p
else:
    return seedEdge               // seedEdge itself is opposite to p
```

In a triangle with edges AB, BC, CA (counter-clockwise), if P is at vertex A (origin of AB), then the opposite edge is BC (AB.Next). If P is at vertex B (origin of BC), the opposite edge is CA. If P is at vertex C (origin of CA), the opposite edge is AB.

---

### `pointInCircumcircle(pt, a, b, c Point) bool`

Checks whether point `pt` lies inside the circumcircle of triangle (A, B, C).

**Algorithm:**

1. Compute the circumcenter (cx, cy) and squared radius r² of (A, B, C) via `GetCircumcircle`.
2. Compute the squared distance from `pt` to the circumcenter: `d² = (pt.x − cx)² + (pt.y − cy)²`.
3. Return `true` if `r² + Epsilon > d²` (point is inside the circumcircle).

The `Epsilon` tolerance ensures points very close to the circumcircle boundary are treated as inside, preventing floating-point oscillation.

---

### `GetCircumcircle(p1, p2, p3 Point) (ux, uy, rSqrd float64, ok bool)`

Computes the circumcenter and squared circumradius of a triangle.

**Math:**

The circumcenter is the intersection of the perpendicular bisectors. Using the standard formula:

```
d = 2 * (p1.x * (p2.y - p3.y) + p2.x * (p3.y - p1.y) + p3.x * (p1.y - p2.y))

ux = (|p1|² * (p2.y - p3.y) + |p2|² * (p3.y - p1.y) + |p3|² * (p1.y - p2.y)) / d
uy = (|p1|² * (p3.x - p2.x) + |p2|² * (p1.x - p3.x) + |p3|² * (p2.x - p1.x)) / d
```

Where `|p|² = p.x² + p.y²`.

The squared radius is the squared distance from the circumcenter to any vertex:
```
r² = (ux - p1.x)² + (uy - p1.y)²
```

**Degenerate case:** If `|d| < Epsilon`, the points are collinear (no valid circumcircle exists) → returns `ok = false`.