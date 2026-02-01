# Algorithm Selection & Justification

## 1. Delaunay Triangulation Strategy

**Selected Approach:** Incremental Insertion with Lawson's Flip (Sloan's Method)

We utilise the iterative insertion method described by Sloan. Our implementation strictly follows **Lawson's swapping algorithm** for restoring the Delaunay property, which is distinct from the cavity-creation approach of Bowyer-Watson.

* **Reference:** Sloan, S. W., "A fast algorithm for constructing Delaunay triangulations in the plane", *Advances in Engineering Software*, 1987.

  * [Semantic Scholar Link](https://www.semanticscholar.org/paper/A-fast-algorithm-for-constructing-Delaunay-in-the-Sloan/ab552a51f2f48af6d17855431c56a71db115c52b)

### 1.1 The Super Triangle

To initialize the algorithm, we construct a "Super Triangle" that encompasses all input points. This ensures the mesh is always a single connected component and simplifies point location logic.

* **Source:** The strategy of creating a bounding triangle significantly larger (e.g., 20x margin) than the input set is a standard technique described in "CAD From Scratch".

  * [CAD From Scratch GitHub & Video Series](https://github.com/xmdi/CAD-from-Scratch)

### 1.2 Point Location (Sloan's Walk)

To locate the triangle containing a new point $P$, we use a "Directed Walk" (or "Walking Search"). Starting from a recently created triangle, we traverse the mesh by moving towards the neighbour that lies in the direction of $P$.

* **Complexity:** $O(\sqrt{N})$ on average for random points.

* **Stochasticity:** We cache the index of the last created triangle (`lastCreated`) to exploit spatial locality of incoming points.

### 1.3 Deviations from Sloan's 1987 Algorithm

While the core logic follows Sloan, the following optimizations described in the paper are adapted:

1. **Bin Sorting:** We implement unidimesional pre-sorting by the X-coordinate. This approximates the spatial locality benefits of full binning, keeping runtime closer to $O(N^{5/4})$.

2. **Coordinate Normalization:** The paper suggests normalizing coordinates to the range $[0, 1]$ to ensure consistent floating-point behavior. Our implementation operates in world-space coordinates.

    * *Impact:* `EPSILON` values in `geometry.go` must be tuned relative to the scale of the map data.

### 1.4 Degeneracy Handling

* **Collinear Points:** If a new point lies directly on an existing edge (within `EPSILON`), the standard algorithm requires an "Edge Split" (1-to-4 split).

* **Current Behavior:** The system detects points on existing edges and performs robust edge splitting as described in Section 3.2. This ensures topological correctness for collinear inputs (e.g., grids).

## 2. Edge Flipping (Lawson's Flip)

**Purpose:** Restore the Delaunay property after insertion.

**Logic:**

After inserting a point and splitting a triangle, we check the quadrilaterals formed by the new triangles and their neighbours. If a pair of triangles shares an edge that is not "locally Delaunay" (i.e., the opposite vertex lies inside the circumcircle), the edge is flipped.

* **Educational Ref:** [GEO1015: Triangulations & Voronoi Diagrams (TU Delft)](https://3d.bk.tudelft.nl/courses/geo1015/)

## 3. Degeneracy Handling: Edge Splitting

### 3.1 Problem Statement

When a new point lies exactly on an existing edge (within EPSILON tolerance), the standard 1-to-3 triangle split fails. This creates a **degenerate case** that requires special handling.

### 3.2 Edge Splitting Algorithm (1-to-4 Split)

When point $P$ lies on the shared edge between triangles $T$ and $N$:

1. **Identify Edge Topology**
   - Shared edge vertices: $u, v$  
   - Opposite vertices: $o$ (in $T$), $o_n$ (in $N$)
   - Original triangles: $T(u,v,o)$ and $N(v,u,o_n)$

2. **Split Triangle $T$**
   - Create $T_1(p,v,o)$ and $T_2(u,p,o)$
   - Maintain CCW orientation from original triangle

3. **Split Triangle $N$** (if not on boundary)
   - Create $N_1(p,u,o_n)$ and $N_2(v,p,o_n)$
   - Maintain CCW orientation from original triangle

4. **Update Neighbor Relationships**
   - Connect new triangles to external neighbours
   - Link the four new triangles appropriately
   - Update boundary pointers (-1 for hull edges)

### 3.3 Neighbor Index Mapping

The algorithm must correctly map existing neighbour indices to new triangles:

**For Triangle $T$:**
- Edge $vo$ (opposite $u$): Neighbor index preserved
- Edge $ou$ (opposite $v$): Neighbor index preserved  
- Edge $uv$ (shared): Connected to new triangle from $N$

**For Triangle $N$:**
- Edge $u o_n$ (opposite $v$): Neighbor index preserved
- Edge $o_n v$ (opposite $u$): Neighbor index preserved
- Edge $vu$ (shared): Connected to new triangle from $T$

### 3.4 Boundary Cases

If the edge lies on the convex hull boundary ($N$ doesn't exist):
- Perform 1-to-2 split instead of 1-to-4
- New triangles on boundary have neighbour index -1 for hull edges

### 3.5 Implementation Notes

* **Edge Detection:** Use `orient2d(pA, pB, p) < EPSILON` to test if point $P$ lies on edge $AB$
* **Triangle Orientation:** Maintain CCW (counter-clockwise) vertex ordering
* **Index Consistency:** Update all neighbour references before legalisation
* **Complexity:** Edge splitting is $O(1)$ per edge, doesn't affect overall $O(N \log N)$ complexity

*Reference: Sloan, S. W., "A fast algorithm for constructing Delaunay triangulations in the plane", Advances in Engineering Software, 1987.*

## 4. Pathfinding Heuristics (Dynamic Fusion)

**Graph Construction:** The mesh is exported as a graph where nodes are triangle centroids and edges represent adjacency.

* **Reference:** Liu, Z., et al., "A Dynamic Fusion Pathfinding Algorithm Using Delaunay Triangulation and Improved A-Star for Mobile Robots", *IEEE Access*, 2021.

  * [ResearchGate Full Text](https://www.researchgate.net/publication/348852130_A_Dynamic_Fusion_Pathfinding_Algorithm_Using_Delaunay_Triangulation_and_Improved_A-star_for_Mobile_Robots_January_2021)

This structure supports **D* Lite** for dynamic replanning when obstacles move, as the local topology repair (edge flipping) minimizes the graph reconstruction cost compared to grid-based methods.

* **Reference:** Koenig, S. & Likhachev, M., "D* Lite", *AAAI/IAAI*, 2002.

  * [AAAI Conference Paper (PDF)](https://aaai.org/Papers/AAAI/2002/AAAI02-072.pdf)