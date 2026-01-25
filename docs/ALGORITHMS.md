# Algorithm Selection & Justification

## 1. Delaunay Triangulation Strategy

**Selected Approach:** Incremental Insertion with Lawson's Flip (Sloan's Method)

We utilize the iterative insertion method described by Sloan. Our implementation strictly follows **Lawson's swapping algorithm** for restoring the Delaunay property, which is distinct from the cavity-creation approach of Bowyer-Watson.

* **Reference:** Sloan, S. W., "A fast algorithm for constructing Delaunay triangulations in the plane", *Advances in Engineering Software*, 1987.

  * [Semantic Scholar Link](https://www.semanticscholar.org/paper/A-fast-algorithm-for-constructing-Delaunay-in-the-Sloan/ab552a51f2f48af6d17855431c56a71db115c52b)

### 1.1 The Super Triangle

To initialize the algorithm, we construct a "Super Triangle" that encompasses all input points. This ensures the mesh is always a single connected component and simplifies point location logic.

* **Source:** The strategy of creating a bounding triangle significantly larger (e.g., 20x margin) than the input set is a standard technique described in "CAD From Scratch".

  * [CAD From Scratch GitHub & Video Series](https://github.com/xmdi/CAD-from-Scratch)

### 1.2 Point Location (Sloan's Walk)

To locate the triangle containing a new point $P$, we use a "Directed Walk" (or "Walking Search"). Starting from a recently created triangle, we traverse the mesh by moving towards the neighbor that lies in the direction of $P$.

* **Complexity:** $O(\sqrt{N})$ on average for random points.

* **Stochasticity:** We cache the index of the last created triangle (`lastCreated`) to exploit spatial locality of incoming points.

### 1.3 Deviations from Sloan's 1987 Algorithm

While the core logic follows Sloan, the following optimizations described in the paper are adapted:

1. **Bin Sorting:** We implement unidimesional pre-sorting by the X-coordinate. This approximates the spatial locality benefits of full binning, keeping runtime closer to $O(N^{5/4})$.

2. **Coordinate Normalization:** The paper suggests normalizing coordinates to the range $[0, 1]$ to ensure consistent floating-point behavior. Our implementation operates in world-space coordinates.

    * *Impact:* `EPSILON` values in `geometry.go` must be tuned relative to the scale of the map data.

### 1.4 Degeneracy Handling

* **Collinear Points:** If a new point lies directly on an existing edge (within `EPSILON`), the standard algorithm requires an "Edge Split" (1-to-4 split).

* **Current Behavior:** The system detects points on existing edges and performs a robust **Edge Split**. The shared edge is split at the new point, dividing the two adjacent triangles into four (or two if on a boundary), ensuring topological correctness even for collinear inputs (e.g., grids).

## 2. Edge Flipping (Lawson's Flip)

**Purpose:** Restore the Delaunay property after insertion.

**Logic:**

After inserting a point and splitting a triangle, we check the quadrilaterals formed by the new triangles and their neighbors. If a pair of triangles shares an edge that is not "locally Delaunay" (i.e., the opposite vertex lies inside the circumcircle), the edge is flipped.

* **Educational Ref:** [GEO1015: Triangulations & Voronoi Diagrams (TU Delft)](https://3d.bk.tudelft.nl/courses/geo1015/)

## 3. Pathfinding Heuristics (Dynamic Fusion)

**Graph Construction:** The mesh is exported as a graph where nodes are triangle centroids and edges represent adjacency.

* **Reference:** Liu, Z., et al., "A Dynamic Fusion Pathfinding Algorithm Using Delaunay Triangulation and Improved A-Star for Mobile Robots", *IEEE Access*, 2021.

  * [ResearchGate Full Text](https://www.researchgate.net/publication/348852130_A_Dynamic_Fusion_Pathfinding_Algorithm_Using_Delaunay_Triangulation_and_Improved_A-star_for_Mobile_Robots_January_2021)

This structure supports **D* Lite** for dynamic replanning when obstacles move, as the local topology repair (edge flipping) minimizes the graph reconstruction cost compared to grid-based methods.

* **Reference:** Koenig, S. & Likhachev, M., "D* Lite", *AAAI/IAAI*, 2002.

  * [AAAI Conference Paper (PDF)](https://aaai.org/Papers/AAAI/2002/AAAI02-072.pdf)