# Mathematical Derivations

## 1. Geometric Predicates

To ensure numerical stability, we avoid calculating angles directly. Instead, we use determinant-based predicates as recommended by Sloan and standard computational geometry literature.

### 1.1 Orientation (Point in Triangle)

To determine if a point $P$ lies inside a triangle $ABC$, we calculate the orientation of the ordered triplets $(A,B,P)$, $(B,C,P)$, and $(C,A,P)$.

* **Formula:** Standard cross-product in 2D (Or "Counter-Clockwise" test).

* **Source:** [Startinpy / GEO1015 Documentation](https://www.researchgate.net/publication/385772992_startinpy_A_Python_library_for_modelling_and_processing_25D_triangulated_terrains)

### 1.2 In-Circle Test

To check if a point $D$ lies inside the circumcircle of triangle $ABC$, we use the determinant of the matrix formed by lifting the points onto a paraboloid $(x, y, x^2 + y^2)$.

$$
det
begin{vmatrix}
 A_x & A_y & A_x^2 + A_y^2 & 1 \
 B_x & B_y & B_x^2 + B_y^2 & 1 \
 C_x & C_y & C_x^2 + C_y^2 & 1 \
 D_x & D_y & D_x^2 + D_y^2 & 1
end{vmatrix}
> 0
$$

* **Why:** This method is robust against floating-point errors compared to calculating the circumcentre explicitly. Explicit circumcentre formulas ($U_x, U_y$) degrade when triangles are "slivers" (nearly collinear).
* **Reference:** Sloan, S. W., "A fast algorithm for constructing Delaunay triangulations in the plane".
  * [Semantic Scholar Link](https://www.semanticscholar.org/paper/A-fast-algorithm-for-constructing-Delaunay-in-the-Sloan/ab552a51f2f48af6d17855431c56a71db115c52b)

## 2. Voronoi Duality

The Voronoi diagram is derived as the dual of the Delaunay triangulation.

* **Relation:** The circumcenter of a Delaunay triangle $T$ becomes a vertex $V$ in the Voronoi diagram. Two Voronoi vertices are connected if their corresponding Delaunay triangles share an edge.

* **Resource:** [GEO1004: Tetrahedralisations and 3D Voronoi diagrams (Video)](https://www.youtube.com/watch?v=oOGx9PUGb5c)

## 3. Circumcentre Calculation

To generate the Voronoi graph (Dual), we calculate the circumcentre $(U_x, U_y)$ of each triangle.

$$ D = 2(A_x(B_y - C_y) + B_x(C_y - A_y) + C_x(A_y - B_y)) $$

$$ U_x = \frac{1}{D} [(A_x^2 + A_y^2)(B_y - C_y) + (B_x^2 + B_y^2)(C_y - A_y) + (C_x^2 + C_y^2)(A_y - B_y)] $$

$$ U_y = \frac{1}{D} [(A_x^2 + A_y^2)(C_x - B_x) + (B_x^2 + B_y^2)(A_x - C_x) + (C_x^2 + C_y^2)(B_x - A_x)] $$

* **Source:** LC4 Notes 1.2.1

## 4. Memory Allocation (Euler's Formula)

For a planar graph with $N$ vertices, the maximum number of triangles is $2N - 2 - k$ (where $k$ is the number of vertices on the convex hull). 

**Optimization:** While Euler's formula gives the theoretical maximum for the *final* mesh, the incremental construction process involves temporary triangle creation and deletion (flipping). To minimize reallocation overhead, we preallocate capacity for $2.5N + 100$ triangles.