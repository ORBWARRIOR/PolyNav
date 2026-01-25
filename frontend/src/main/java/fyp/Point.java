package fyp;
import java.util.List;

/**
 * Point class to represent a point in 2D space
 */
public class Point {

    double x;
    double y;
    private List<Point> neighbours;
    private List<Triangle> triangles;

    Point(double x, double y) {
        this.x = x;
        this.y = y;
    }

    public boolean addNeighbour(Point neighbour) {
        if (!neighbours.contains(neighbour)) {
            return neighbours.add(neighbour);
        }
        return false;
    }

    public boolean removeNeighbour(Point neighbour) {
        if (neighbours.contains(neighbour)) {
            return neighbours.remove(neighbour);
        }
        return false;
    }

    public boolean addTriangle(Triangle triangle) {
        if (!triangles.contains(triangle)) {
            return triangles.add(triangle);
        }
        return false;
    }

    public boolean removeTriangle(Triangle triangle) {
        if (triangles.contains(triangle)) {
            return triangles.remove(triangle);
        }
        return false;
    }
}
