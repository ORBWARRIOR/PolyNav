# Algorithm Implementation Guide

This document explains the structure and responsibility of each source file within the `backend/internal/algo/` directory. It complements the theoretical overviews in [ALGORITHMS.md](ALGORITHMS.md) and [MATHEMATICS.md](MATHEMATICS.md).

## File Overview

### 1. `types.go`

**Role:** Data Structure Definitions
Defines the core data structures used throughout the package.

* **`Point`**: Basic 2D coordinate $(X, Y)$.
* **`Triangle`**: Stores 3 vertex indices (`A, B, C`) and 3 neighbor indices (`T1, T2, T3`).
* *Convention:* Neighbor `T1` is the triangle sharing the edge opposite vertex `A` (edge `BC`).
* *Flag:* `Active` allows for logical deletion without array resizing.
* **`Delaunay`**: The main context struct holding the mesh state (slices of Points and Triangles) and acceleration structures.

### 2. `delaunay.go`

**Role:** Initialization & Lifecycle
Handles the setup and high-level execution flow of the triangulation.

* **`NewDelaunay`**: Initializes the mesh with a "Super Triangle".
* *Note:* It calculates the bounding box of input points to size the Super Triangle (20x margin) but **does not normalize** the points to a unit square.
* **`Triangulate`**: The main driver function. It iterates through all input points and calls `insertPoint` for each.
* **`cleanup`**: Removes the Super Triangle vertices and any triangles attached to them after triangulation is complete.

### 3. `insertions.go`
 
 **Role:** Core Algorithm Logic (Sloan's Method)
 Implements the "Incremental Insertion" algorithm.
 
 * **`insertPoint`**: Adds a single point to the mesh.
     1. **Locate:** Finds the enclosing triangle using `walkLocate`.
     2. **Degeneracy Check:** Checks if the point lies on an edge (using `orient2d < EPSILON`).
         * *Action:* If collinear, calls `splitEdge` to perform a topological split.
     3. **Split:** Splits the enclosing triangle into three new ones (standard 1-to-3).
     4. **Legalize:** Recursively calls `legaliseEdge` to flip edges that violate the Delaunay condition.
 * **`splitEdge`**: Handles the degenerate case where a point falls on an existing edge. It splits the two triangles sharing that edge into four (or two for boundary edges) to maintain a valid mesh.
 * **`walkLocate`**: Implements Sloan's "Directed Walk" to find the triangle containing a query point.
 * **`legaliseEdge`**: Performs Lawson's Edge Flip. If a point lies inside the circumcircle of an adjacent triangle, the shared edge is flipped.

### 4. `geometry.go`

**Role:** Mathematical Predicates
Contains the low-level geometric calculations required for robust decision making.

* **`orient2d`**: Determines if a point is to the left, right, or on the line defined by two other points (Cross Product).
* **`inCircumcircle`**: The "In-Circle" test using a determinant-based approach (lifting points to a paraboloid).
* **`contains`**: Helper to check if a point is strictly inside a triangle using orientation tests.

### 5. `graph.go`

**Role:** Dual Graph Generation
Handles the conversion of the triangular mesh into a graph structure suitable for pathfinding algorithms like A*.

* **`ExportGraph`**: Calculates the circumcenter of every active triangle. These circumcenters become the nodes of the Voronoi graph.

### 6. `debug.go`

**Role:** Visualization & Debugging

* **`DebugJSON`**: Serializes the current state of the mesh into **GeoJSON** format for external visualization.

### 7. `delaunay_test.go`

**Role:** Testing & Benchmarking

* **`TestDelaunay_Triangulate_Scenarios`**: Table-driven tests covering standard cases.
* **`TestDelaunay_Random_Stress`**: Verifies stability with large random datasets.
