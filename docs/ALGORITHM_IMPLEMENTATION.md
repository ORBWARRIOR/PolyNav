# Algorithm Implementation Guide

This document explains the structure and responsibility of each source file within the `backend/internal/algo/` directory. It complements the theoretical overviews in [ALGORITHMS.md](ALGORITHMS.md) and [MATHEMATICS.md](MATHEMATICS.md).

## File Overview

### 1. `types.go`
**Role:** Data Structure Definitions
Defines the core data structures used throughout the package.
*   **`Point`**: Basic 2D coordinate $(X, Y)$.
*   **`Triangle`**: Stores 3 vertex indices (`A, B, C`) and 3 neighbor indices (`T1, T2, T3`).
    *   *Convention:* Neighbor `T1` is the triangle sharing the edge opposite vertex `A` (edge `BC`).
    *   *Flag:* `Active` allows for logical deletion without array resizing.
*   **`Delaunay`**: The main context struct holding the mesh state (slices of Points and Triangles) and acceleration structures (e.g., `lastCreated` index for search caching).
*   **`GraphNode`**: Represents a node in the Dual Graph (Voronoi), used for exporting the mesh to pathfinding algorithms.

### 2. `delaunay.go`
**Role:** Initialization & Lifecycle
Handles the setup and high-level execution flow of the triangulation.
*   **`NewDelaunay`**: Initializes the mesh with a "Super Triangle" (as described in *ALGORITHMS.md*) to ensure all points are enclosed within a convex hull initially.
*   **`Triangulate`**: The main driver function. It iterates through all input points and calls `insertPoint` for each.
*   **`cleanup`**: Removes the Super Triangle vertices and any triangles attached to them after triangulation is complete, leaving only the triangulation of the actual input points.

### 3. `insertions.go`
**Role:** Core Algorithm Logic (Sloan's Method)
Implements the "Incremental Insertion" algorithm.
*   **`insertPoint`**: Adds a single point to the mesh. It:
    1.  Finds the enclosing triangle using `walkLocate`.
    2.  Splits that triangle into three new ones.
    3.  Recursively calls `legaliseEdge` to flip edges that violate the Delaunay condition.
*   **`walkLocate`**: Implements Sloan's "Directed Walk" to find the triangle containing a query point. This is an optimization over linear search ($O(\sqrt{N})$ vs $O(N)$).
*   **`legaliseEdge`**: Performs Lawson's Edge Flip. If a point lies inside the circumcircle of an adjacent triangle, the shared edge is flipped to maximize the minimum angle (restoring the Delaunay property).

### 4. `geometry.go`
**Role:** Mathematical Predicates
Contains the low-level geometric calculations required for robust decision making.
*   **`orient2d`**: Determines if a point is to the left, right, or on the line defined by two other points (Cross Product).
*   **`inCircumcircle`**: The "In-Circle" test. Determines if a point lies inside the circumcircle of a triangle. Uses a determinant-based approach (lifting points to a paraboloid) for numerical stability (see *MATHEMATICS.md*).
*   **`contains`**: Helper to check if a point is strictly inside a triangle using orientation tests.

### 5. `graph.go`
**Role:** Dual Graph Generation
Handles the conversion of the triangular mesh into a graph structure suitable for pathfinding algorithms like A*.
*   **`ExportGraph`**: Calculates the circumcenter of every active triangle. These circumcenters become the nodes of the Voronoi graph. It links nodes based on triangle adjacency.

### 6. `debug.go`
**Role:** Visualization & Debugging
*   **`DebugJSON`**: Serializes the current state of the mesh into **GeoJSON** format.
    *   *Usage:* The output can be pasted directly into tools like [geojson.io](https://geojson.io) to visually inspect the mesh, check for holes, or verify neighbor connections.

### 7. `delaunay_test.go`
**Role:** Testing & Benchmarking
*   **`TestDelaunay_Triangulate_Scenarios`**: Table-driven tests covering standard cases (squares, grids) and edge cases (collinear points, degenerate triangles).
*   **`TestDelaunay_Random_Stress`**: Verifies stability with a large number of randomly generated points (1000+) using deterministic seeds.
*   **`BenchmarkDelaunay`**: Measures the performance of the triangulation algorithm to detect regressions.