# Algorithm Selection & Justification

## 1. Delaunay Triangulation Strategy

**Selected Approach:** Incremental Insertion (Bowyer-Watson variant)

We selected the incremental insertion method over divide-and-conquer to satisfy the requirement for dynamic updates in pathfinding scenarios.

* **Reference:** Sloan, S. W., "A fast algorithm for constructing Delaunay triangulations in the plane", *Advances in Engineering Software*, 1987.
  * [Semantic Scholar Link](https://www.semanticscholar.org/paper/A-fast-algorithm-for-constructing-Delaunay-in-the-Sloan/ab552a51f2f48af6d17855431c56a71db115c52b)

### 1.1 The Super Triangle

To initialize the algorithm, we construct a "Super Triangle" that encompasses all input points. This ensures the mesh is always a single connected component and simplifies point location logic.

* **Source:** The strategy of creating a bounding triangle significantly larger (e.g., 20x margin) than the input set is a standard technique described in "CAD From Scratch".
  * [CAD From Scratch GitHub & Video Series](https://github.com/xmdi/CAD-from-Scratch)

### 1.2 Point Location (Sloan's Walk)

To locate the triangle containing a new point $P$, we use a "Directed Walk" (or "Walking Search"). Starting from a recently created triangle, we traverse the mesh by moving towards the neighbor that lies in the direction of $P$.

*   **Complexity:** $O(\sqrt{N})$ on average for random points, compared to $O(N)$ for linear scan.
*   **Stochasticity:** We cache the index of the last created triangle (`lastCreated`) to exploit spatial locality of incoming points.

## 2. Edge Flipping (Lawson's Flip)

**Purpose:** Restore the Delaunay property after insertion.

**Logic:**
After inserting a point and splitting a triangle, we check the quadrilaterals formed by the new triangles and their neighbors. If a pair of triangles shares an edge that is not "locally Delaunay" (i.e., the opposite vertex lies inside the circumcircle), the edge is flipped. This local optimization is critical for maintaining the empty-circle property required for high-quality meshes.

* **Educational Ref:** [GEO1015: Triangulations & Voronoi Diagrams (TU Delft)](https://3d.bk.tudelft.nl/courses/geo1015/)

## 3. Pathfinding Heuristics (Dynamic Fusion)

**Graph Construction:** The mesh is exported as a graph where nodes are triangle centroids and edges represent adjacency.

* **Reference:** Liu, Z., et al., "A Dynamic Fusion Pathfinding Algorithm Using Delaunay Triangulation and Improved A-Star for Mobile Robots", *IEEE Access*, 2021.
  * [ResearchGate Full Text](https://www.researchgate.net/publication/348852130_A_Dynamic_Fusion_Pathfinding_Algorithm_Using_Delaunay_Triangulation_and_Improved_A-star_for_Mobile_Robots_January_2021)

This structure supports **D* Lite** for dynamic replanning when obstacles move, as the local topology repair (edge flipping) minimizes the graph reconstruction cost compared to grid-based methods.

* **Reference:** Koenig, S. & Likhachev, M., "D* Lite", *AAAI/IAAI*, 2002.

  * [AAAI Conference Paper (PDF)](https://aaai.org/Papers/AAAI/2002/AAAI02-072.pdf)
